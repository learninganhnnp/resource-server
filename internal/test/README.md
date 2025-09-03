# Resource Server E2E Test Suite

This directory contains the comprehensive End-to-End (E2E) test suite for the Resource Server API. The tests validate all API endpoints, ensure proper integration between components, and verify business logic workflows.

## Overview

The E2E test suite covers:
- All API endpoints with various scenarios
- Error handling and validation
- Complex workflow tests
- Cross-provider operations  
- Performance and concurrency tests

## Test Structure

```
test/
├── e2e/                     # E2E test files
│   ├── suite_test.go        # Base test suite setup
│   ├── health_test.go       # Health check endpoint tests
│   ├── resource_definition_test.go  # Resource definition tests
│   ├── provider_test.go     # Provider endpoint tests
│   ├── file_operations_test.go      # File operation tests
│   ├── multipart_test.go    # Multipart upload tests
│   ├── achievement_test.go  # Achievement API tests
│   └── workflows_test.go    # Complex workflow tests
├── fixtures/                # Test data and fixtures
│   ├── achievements.json    # Sample achievement data
│   ├── metadata.json        # Test metadata and parameters
│   └── files/              # Test files for upload
├── helpers/                 # Test helper functions
│   ├── client.go           # API client helpers
│   ├── database.go         # Database setup/teardown
│   ├── assertions.go       # Custom test assertions
│   └── mocks.go           # Mock storage providers
└── README.md              # This file
```

## Prerequisites

### Required Software
- Go 1.24.2+
- PostgreSQL 15+
- Docker and Docker Compose (optional)

### Environment Variables
- `TEST_DATABASE_DSN` - PostgreSQL connection string for tests
- `SERVER_PORT` - Port for test server (default: 8081)

## Running Tests

### Option 1: Using Make (Recommended)

```bash
# Run all tests
make test-all

# Run only E2E tests
make test-e2e

# Run with coverage
make coverage

# Run workflow tests only
make test-workflows

# Quick development test
make dev-test
```

### Option 2: Using the Test Runner Script

```bash
# Run all E2E tests
./scripts/run-e2e-tests.sh run-all

# Run specific test suite
./scripts/run-e2e-tests.sh run-suite -s health
./scripts/run-e2e-tests.sh run-suite -s workflows

# Use Docker for database
./scripts/run-e2e-tests.sh --docker run-all

# Custom configuration
./scripts/run-e2e-tests.sh -p 8082 -t 600 run-all
```

### Option 3: Using Go Test Directly

```bash
# Start test database
make setup-test-db

# Build and start server
go build -o resource-server ./cmd/api
./resource-server &

# Run tests
go test -v -tags=e2e ./internal/test/e2e/...

# Cleanup
kill %1
make teardown-test-db
```

### Option 4: Using Docker Compose

```bash
# Start test environment
docker-compose -f docker-compose.test.yml up -d

# Run tests in container
docker-compose -f docker-compose.test.yml exec resource-server-test \
  go test -v -tags=e2e ./internal/test/e2e/...

# Cleanup
docker-compose -f docker-compose.test.yml down
```

## Test Suites

### Health Check Tests (`health_test.go`)
- **HC-001**: Successful health check
- **HC-002**: Verify response format
- **HC-003**: Check content-type header
- **HC-004**: Verify server readiness

### Resource Definition Tests (`resource_definition_test.go`)
- **RD-001-004**: List all definitions with structure validation
- **RD-005-008**: Get definition by name with error cases

### Provider Tests (`provider_test.go`)
- **PR-001-007**: List and get providers with capability validation

### File Operations Tests (`file_operations_test.go`)
- **FO-001-033**: Complete file lifecycle operations
  - List files with filters and pagination
  - Generate upload/download URLs
  - Metadata operations
  - File deletion

### Multipart Upload Tests (`multipart_test.go`)
- **MP-001-011**: Multipart upload workflow
  - Initialize upload
  - Generate part URLs
  - Complete/abort operations

### Achievement Tests (`achievement_test.go`)
- **AC-001-027**: Complete achievement management
  - CRUD operations
  - Icon upload workflow
  - Validation and error handling

### Workflow Tests (`workflows_test.go`)
Complex end-to-end scenarios:
1. **Complete Achievement Creation with Icon Upload**
2. **Large File Multipart Upload**
3. **File Lifecycle Management**
4. **Achievement Management Flow**
5. **Cross-Provider Operations**

## Test Data

### Fixtures
- `achievements.json` - 12 sample achievements with various categories
- `metadata.json` - Test parameters, scopes, and provider configurations
- `files/` - Binary test files for upload scenarios

### Test Parameters
- Achievement IDs: UUIDs for testing
- Scopes: Global (G), App (A), ClientApp (CA)
- Providers: CDN, GCS, R2
- File formats: PNG, JPEG, WebP

## Configuration

### Database Setup
The test suite requires a PostgreSQL database. Configure with:

```bash
export TEST_DATABASE_DSN="postgres://user:pass@localhost:5432/resource_test?sslmode=disable"
```

Or use Docker:
```bash
make setup-test-db
```

### Provider Configuration
Tests will skip provider-specific operations if providers are not configured or available.

## Coverage Requirements

The test suite aims for:
- API endpoint coverage: 100%
- Business logic coverage: >80%
- Error handling coverage: >90%
- Integration workflow coverage: 100%

## CI/CD Integration

### GitHub Actions
The `.github/workflows/e2e-tests.yml` workflow runs:
- Unit tests
- E2E tests
- Coverage reporting
- Security scanning
- Performance baselines

### Local CI Simulation
```bash
make ci
```

## Troubleshooting

### Common Issues

**Database Connection Failed**
```bash
# Check PostgreSQL is running
pg_isready -h localhost -p 5432

# Or use Docker
make setup-test-db
```

**Server Start Timeout**
```bash
# Check if port is already in use
lsof -i :8081

# Use different port
export SERVER_PORT=8082
```

**Tests Fail Due to Missing Dependencies**
```bash
# Install dependencies
go mod download
go mod tidy
```

**Docker Issues**
```bash
# Clean up Docker resources
docker-compose -f docker-compose.test.yml down -v
docker system prune
```

### Debug Mode
Enable verbose logging:
```bash
export VERBOSE=true
./scripts/run-e2e-tests.sh run-all
```

### Test Specific Endpoints
```bash
# Test only health endpoint
go test -v -tags=e2e -run=TestHealthSuite ./internal/test/e2e/health_test.go

# Test specific workflow
go test -v -tags=e2e -run=TestWorkflow_AchievementWithIcon ./internal/test/e2e/workflows_test.go
```

## Contributing

### Adding New Tests
1. Create test file in `test/e2e/`
2. Follow existing naming conventions (TestCase-ID format)
3. Add test data to fixtures if needed
4. Update this README

### Test Guidelines
- Use descriptive test names
- Include test case IDs for tracking
- Add proper cleanup in `SetupTest`/`TearDownTest`
- Mock external dependencies
- Test both success and error scenarios

### Performance Considerations
- Keep test data minimal
- Use parallel tests where possible
- Clean up resources promptly
- Monitor test execution times

## Maintenance

### Regular Tasks
- Update test data for API changes
- Review and update mock responses
- Monitor test execution times
- Investigate flaky tests
- Update provider configurations

### Test Review Checklist
- [ ] All new endpoints have tests
- [ ] Error cases are covered
- [ ] Performance impact assessed
- [ ] Documentation updated
- [ ] CI pipeline passing
- [ ] No hardcoded values
- [ ] Proper cleanup implemented

## References

- [Go Testing Package](https://pkg.go.dev/testing)
- [Testify Framework](https://github.com/stretchr/testify)
- [API Documentation](../api-docs/)
- [Contributing Guidelines](../CONTRIBUTING.md)