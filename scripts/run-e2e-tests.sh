#!/bin/bash

# E2E Test Runner Script
# This script sets up the test environment and runs end-to-end tests

set -e

# Default values
DATABASE_DSN=${TEST_DATABASE_DSN:-"postgres://postgres:anhnguyen0809@localhost:5432/resource_test?sslmode=disable"}
SERVER_PORT=${SERVER_PORT:-8081}
TEST_TIMEOUT=${TEST_TIMEOUT:-300}
COVERAGE=${COVERAGE:-true}
VERBOSE=${VERBOSE:-true}
USE_DOCKER=${USE_DOCKER:-false}

# Environment variables for the application
DB_NAME=${DB_NAME:-"resource_test"}
DB_USERNAME=${DB_USERNAME:-"postgres"}
DB_PASSWORD=${DB_PASSWORD:-"anhnguyen0809"}
DB_HOST=${DB_HOST:-"localhost"}
DB_PORT=${DB_PORT:-"5432"}
DB_SSLMODE=${DB_SSLMODE:-"disable"}

# Colors for output
RED='\033[0;31m'
GREEN='\033[0;32m'
YELLOW='\033[1;33m'
BLUE='\033[0;34m'
NC='\033[0m' # No Color

# Logging functions
log_info() {
    echo -e "${BLUE}[INFO]${NC} $1"
}

log_success() {
    echo -e "${GREEN}[SUCCESS]${NC} $1"
}

log_warning() {
    echo -e "${YELLOW}[WARNING]${NC} $1"
}

log_error() {
    echo -e "${RED}[ERROR]${NC} $1"
}

# Cleanup function
cleanup() {
    local exit_code=$?
    log_info "Cleaning up..."
    
    if [[ -f server.pid ]]; then
        local pid=$(cat server.pid)
        if kill -0 $pid 2>/dev/null; then
            log_info "Stopping server (PID: $pid)"
            kill $pid
            wait $pid 2>/dev/null || true
        fi
        rm -f server.pid
    fi
    
    if [[ "$USE_DOCKER" == "true" ]]; then
        log_info "Stopping Docker services"
        docker-compose -f docker-compose.test.yml down
    fi
    
    exit $exit_code
}

# Set trap for cleanup
trap cleanup EXIT INT TERM

# Function to setup environment variables
setup_environment() {
    log_info "Setting up environment variables..."
    
    # Export database configuration
    export DB_NAME="$DB_NAME"
    export DB_USERNAME="$DB_USERNAME" 
    export DB_PASSWORD="$DB_PASSWORD"
    export DB_HOST="$DB_HOST"
    export DB_PORT="$DB_PORT"
    export DB_SSLMODE="$DB_SSLMODE"
    
    # Export test-specific variables
    export TEST_DATABASE_DSN="$DATABASE_DSN"
    export SERVER_PORT="$SERVER_PORT"
    
    # Optional: Load .env file if it exists
    if [[ -f .env ]]; then
        log_info "Loading .env file..."
        export $(grep -v '^#' .env | xargs)
    fi
    
    # Optional: Load .env.test file if it exists (test-specific overrides)
    if [[ -f .env.test ]]; then
        log_info "Loading .env.test file..."
        export $(grep -v '^#' .env.test | xargs)
    fi
    
    log_success "Environment variables configured"
}

# Function to check if a service is running
check_service() {
    local service=$1
    local port=$2
    local max_attempts=${3:-30}
    local attempt=0
    
    log_info "Waiting for $service to be ready on port $port..."
    
    while [[ $attempt -lt $max_attempts ]]; do
        if curl -f http://localhost:$port/api/v1/health >/dev/null 2>&1; then
            log_success "$service is ready!"
            return 0
        fi
        
        attempt=$((attempt + 1))
        sleep 1
    done
    
    log_error "$service failed to start within $max_attempts seconds"
    return 1
}

# Function to run database migrations
run_migrations() {
    local migrations_dir="./migrations"
    
    if [[ ! -d "$migrations_dir" ]]; then
        log_warning "Migrations directory not found at $migrations_dir"
        return 0
    fi
    
    log_info "Running database migrations from $migrations_dir..."
    
    # Get all .up.sql files and sort them
    local migration_files=($(find "$migrations_dir" -name "*.up.sql" | sort))
    
    if [[ ${#migration_files[@]} -eq 0 ]]; then
        log_warning "No migration files found in $migrations_dir"
        return 0
    fi
    
    for migration_file in "${migration_files[@]}"; do
        local migration_name=$(basename "$migration_file")
        log_info "Applying migration: $migration_name"
        
        if [[ "$USE_DOCKER" == "true" ]]; then
            # Run migration via docker exec
            if ! docker-compose -f docker-compose.test.yml exec -T postgres-test psql -U postgres -d resource_test -f "/migrations/$(basename "$migration_file")" >/dev/null 2>&1; then
                log_error "Failed to apply migration: $migration_name"
                return 1
            fi
        else
            # Run migration directly with psql
            if ! psql "$DATABASE_DSN" -f "$migration_file" >/dev/null 2>&1; then
                log_error "Failed to apply migration: $migration_name"
                return 1
            fi
        fi
        
        log_success "Migration applied: $migration_name"
    done
    
    log_success "All migrations completed successfully"
}

# Function to setup database
setup_database() {
    if [[ "$USE_DOCKER" == "true" ]]; then
        log_info "Using Docker for database setup"
        docker-compose -f docker-compose.test.yml up -d postgres-test
        
        # Wait for PostgreSQL to be ready
        local max_attempts=30
        local attempt=0
        
        log_info "Waiting for PostgreSQL to be ready..."
        while [[ $attempt -lt $max_attempts ]]; do
            if docker-compose -f docker-compose.test.yml exec -T postgres-test pg_isready -U postgres -d resource_test >/dev/null 2>&1; then
                log_success "PostgreSQL is ready!"
                break
            fi
            
            attempt=$((attempt + 1))
            sleep 1
        done
        
        if [[ $attempt -eq $max_attempts ]]; then
            log_error "PostgreSQL failed to start"
            return 1
        fi
        
        # Run migrations after database is ready
        run_migrations
        
    else
        log_info "Checking local PostgreSQL connection"
        if ! psql "$DATABASE_DSN" -c "SELECT 1;" >/dev/null 2>&1; then
            log_error "Cannot connect to PostgreSQL. Please ensure it's running."
            log_info "Try: make setup-test-db"
            return 1
        fi
        log_success "PostgreSQL connection verified"
        
        # Run migrations for local database
        run_migrations
    fi
}

# Function to build application
build_application() {
    log_info "Building application..."
    
    if ! go build -o resource-server ./cmd/api; then
        log_error "Failed to build application"
        return 1
    fi
    
    log_success "Application built successfully"
}

# Function to start server
start_server() {
    log_info "Starting server on port $SERVER_PORT..."
    
    ./resource-server &
    local server_pid=$!
    echo $server_pid > server.pid
    
    log_info "Server started with PID: $server_pid"
    
    if ! check_service "Resource Server" "$SERVER_PORT" 30; then
        return 1
    fi
    
    return 0
}

# Function to run tests
run_tests() {
    local test_args=""
    
    if [[ "$VERBOSE" == "true" ]]; then
        test_args="$test_args -v"
    fi
    
    if [[ "$COVERAGE" == "true" ]]; then
        test_args="$test_args -coverprofile=e2e_coverage.out"
    fi
    
    log_info "Running E2E tests..."
    
    local test_cmd="go test $test_args -race -tags=e2e -timeout=${TEST_TIMEOUT}s ./internal/test/e2e/..."
    
    log_info "Running: $test_cmd"
    
    if eval $test_cmd; then
        log_success "E2E tests passed!"
        return 0
    else
        log_error "E2E tests failed!"
        return 1
    fi
}

# Function to run specific test suites
run_test_suite() {
    local suite=$1
    
    case $suite in
        "health")
            go test -v -tags=e2e -run=TestHealthSuite ./internal/test/e2e/health_test.go
            ;;
        "definitions")
            go test -v -tags=e2e -run=TestResourceDefinitionSuite ./internal/test/e2e/resource_definition_test.go
            ;;
        "providers")
            go test -v -tags=e2e -run=TestProviderSuite ./internal/test/e2e/provider_test.go
            ;;
        "files")
            go test -v -tags=e2e -run=TestFileOperationsSuite ./internal/test/e2e/file_operations_test.go
            ;;
        "multipart")
            go test -v -tags=e2e -run=TestMultipartSuite ./internal/test/e2e/multipart_test.go
            ;;
        "achievements")
            go test -v -tags=e2e -run=TestAchievementSuite ./internal/test/e2e/achievement_test.go
            ;;
        "workflows")
            go test -v -tags=e2e -run=TestWorkflowSuite ./internal/test/e2e/workflows_test.go
            ;;
        *)
            log_error "Unknown test suite: $suite"
            return 1
            ;;
    esac
}

# Function to generate coverage report
generate_coverage_report() {
    if [[ "$COVERAGE" == "true" && -f e2e_coverage.out ]]; then
        log_info "Generating coverage report..."
        
        go tool cover -html=e2e_coverage.out -o e2e_coverage.html
        go tool cover -func=e2e_coverage.out
        
        log_success "Coverage report generated: e2e_coverage.html"
    fi
}

# Help function
show_help() {
    cat << EOF
E2E Test Runner

Usage: $0 [OPTIONS] [COMMAND]

Commands:
    run-all         Run all E2E tests (default)
    run-suite       Run specific test suite
    build           Build application only
    setup-db        Setup database only
    migrate         Run database migrations only

Options:
    -d, --database DSN    Database connection string
    -p, --port PORT      Server port (default: 8081)
    -t, --timeout SEC    Test timeout in seconds (default: 300)
    -c, --coverage       Generate coverage report (default: true)
    -v, --verbose        Verbose output (default: true)
    --docker            Use Docker for database (default: false)
    -s, --suite SUITE    Test suite to run (health|definitions|providers|files|multipart|achievements|workflows)
    -h, --help          Show this help

Examples:
    $0 run-all
    $0 run-suite -s health
    $0 --docker run-all
    $0 -p 8082 -t 600 run-all
    $0 migrate
    $0 --docker migrate

Environment Variables:
    TEST_DATABASE_DSN   Database connection string
    SERVER_PORT         Server port
    TEST_TIMEOUT        Test timeout in seconds
    COVERAGE           Enable coverage (true|false)
    VERBOSE            Enable verbose output (true|false)
    USE_DOCKER         Use Docker for services (true|false)
    
    Database Configuration:
    DB_NAME            Database name (default: resource_test)
    DB_USERNAME        Database username (default: postgres)
    DB_PASSWORD        Database password (default: anhnguyen0809) # cspell:disable-line
    DB_HOST            Database host (default: localhost)
    DB_PORT            Database port (default: 5432)
    DB_SSLMODE         Database SSL mode (default: disable) # cspell:disable-line
EOF
}

# Parse command line arguments
parse_args() {
    while [[ $# -gt 0 ]]; do
        case $1 in
            -d|--database)
                DATABASE_DSN="$2"
                shift 2
                ;;
            -p|--port)
                SERVER_PORT="$2"
                shift 2
                ;;
            -t|--timeout)
                TEST_TIMEOUT="$2"
                shift 2
                ;;
            -c|--coverage)
                COVERAGE="$2"
                shift 2
                ;;
            -v|--verbose)
                VERBOSE="$2"
                shift 2
                ;;
            --docker)
                USE_DOCKER="true"
                shift
                ;;
            -s|--suite)
                TEST_SUITE="$2"
                shift 2
                ;;
            -h|--help)
                show_help
                exit 0
                ;;
            run-all)
                COMMAND="run-all"
                shift
                ;;
            run-suite)
                COMMAND="run-suite"
                shift
                ;;
            build)
                COMMAND="build"
                shift
                ;;
            setup-db)
                COMMAND="setup-db"
                shift
                ;;
            migrate)
                COMMAND="migrate"
                shift
                ;;
            *)
                log_error "Unknown option: $1"
                show_help
                exit 1
                ;;
        esac
    done
}

# Main function
main() {
    local command=${COMMAND:-"run-all"}
    
    log_info "Starting E2E test runner..."
    
    # Setup environment variables first
    setup_environment
    
    log_info "Database DSN: $DATABASE_DSN"
    log_info "Server Port: $SERVER_PORT"
    log_info "Use Docker: $USE_DOCKER"
    
    case $command in
        "build")
            build_application
            ;;
        "setup-db")
            setup_database
            ;;
        "migrate")
            run_migrations
            ;;
        "run-suite")
            if [[ -z "$TEST_SUITE" ]]; then
                log_error "Test suite not specified. Use -s|--suite option."
                exit 1
            fi
            
            setup_database
            build_application
            start_server
            run_test_suite "$TEST_SUITE"
            ;;
        "run-all")
            setup_database
            build_application
            start_server
            run_tests
            generate_coverage_report
            ;;
        *)
            log_error "Unknown command: $command"
            show_help
            exit 1
            ;;
    esac
    
    log_success "E2E test runner completed successfully!"
}

# Parse arguments and run
parse_args "$@"
main