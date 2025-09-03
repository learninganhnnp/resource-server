# Makefile for Resource Server E2E Tests

.PHONY: test test-unit test-e2e test-all coverage clean setup-test-db run-tests

# Go parameters
GOCMD=go
GOBUILD=$(GOCMD) build
GOCLEAN=$(GOCMD) clean
GOTEST=$(GOCMD) test
GOGET=$(GOCMD) get
BINARY_NAME=resource-server
BINARY_UNIX=$(BINARY_NAME)_unix
TEST_DATABASE_DSN?=postgres://postgres:password@localhost:5432/resource_test?sslmode=disable

# Build the application
build:
	$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/api

# Clean build artifacts
clean:
	$(GOCLEAN)
	rm -f $(BINARY_NAME)
	rm -f $(BINARY_UNIX)
	rm -f *_coverage.out
	rm -f total_coverage.out
	rm -f coverage.html

# Install dependencies
deps:
	$(GOGET) -v -d ./...
	$(GOCMD) mod download
	$(GOCMD) mod tidy

# Run unit tests
test-unit:
	$(GOTEST) -v -race -coverprofile=unit_coverage.out ./internal/...

# Setup test database
setup-test-db:
	@echo "Setting up test database..."
	@docker run --name postgres-test -e POSTGRES_DB=resource_test -e POSTGRES_PASSWORD=password -p 5432:5432 -d postgres:15 || true
	@sleep 5
	@docker exec postgres-test psql -U postgres -d resource_test -c "CREATE EXTENSION IF NOT EXISTS \"uuid-ossp\";" || true

# Teardown test database
teardown-test-db:
	@echo "Tearing down test database..."
	@docker stop postgres-test || true
	@docker rm postgres-test || true

# Run E2E tests
test-e2e: build
	@echo "Running E2E tests..."
	@export TEST_DATABASE_DSN="$(TEST_DATABASE_DSN)" && \
	./$(BINARY_NAME) & \
	SERVER_PID=$$! && \
	echo "Started server with PID $$SERVER_PID" && \
	sleep 5 && \
	timeout 30 bash -c 'until curl -f http://localhost:8081/api/v1/health; do sleep 1; done' && \
	$(GOTEST) -v -race -coverprofile=e2e_coverage.out -tags=e2e ./test/e2e/... ; \
	TEST_RESULT=$$? && \
	kill $$SERVER_PID && \
	exit $$TEST_RESULT

# Run workflow tests only
test-workflows: build
	@echo "Running workflow tests..."
	@export TEST_DATABASE_DSN="$(TEST_DATABASE_DSN)" && \
	./$(BINARY_NAME) & \
	SERVER_PID=$$! && \
	sleep 5 && \
	timeout 30 bash -c 'until curl -f http://localhost:8081/api/v1/health; do sleep 1; done' && \
	$(GOTEST) -v -race -tags=e2e ./test/e2e/workflows_test.go ; \
	TEST_RESULT=$$? && \
	kill $$SERVER_PID && \
	exit $$TEST_RESULT

# Run all tests
test-all: test-unit test-e2e

# Run tests with coverage
coverage: test-all
	@which gocovmerge || go install github.com/wadey/gocovmerge@latest
	@gocovmerge unit_coverage.out e2e_coverage.out > total_coverage.out
	@$(GOCMD) tool cover -html=total_coverage.out -o coverage.html
	@$(GOCMD) tool cover -func=total_coverage.out
	@echo "Coverage report generated: coverage.html"

# Run short tests only
test-short:
	$(GOTEST) -v -short ./...

# Run tests with race detection
test-race:
	$(GOTEST) -v -race ./...

# Run benchmarks
benchmark: build
	@export TEST_DATABASE_DSN="$(TEST_DATABASE_DSN)" && \
	./$(BINARY_NAME) & \
	SERVER_PID=$$! && \
	sleep 5 && \
	$(GOTEST) -v -bench=. -benchmem -run=^$$ -tags=e2e,performance ./test/e2e/... && \
	kill $$SERVER_PID

# Lint code
lint:
	@which golangci-lint || curl -sSfL https://raw.githubusercontent.com/golangci/golangci-lint/master/install.sh | sh -s -- -b $$(go env GOPATH)/bin v1.50.1
	golangci-lint run

# Format code
fmt:
	$(GOCMD) fmt ./...

# Vet code
vet:
	$(GOCMD) vet ./...

# Security scan
security:
	@which gosec || go install github.com/securecodewarrior/gosec/v2/cmd/gosec@latest
	gosec ./...

# Run CI pipeline locally
ci: deps lint vet security test-all coverage

# Quick development test (unit tests + health check)
dev-test: test-unit
	@echo "Running quick health check..."
	@$(GOBUILD) -o $(BINARY_NAME) -v ./cmd/api
	@./$(BINARY_NAME) & \
	SERVER_PID=$$! && \
	sleep 2 && \
	curl -f http://localhost:8081/api/v1/health && \
	kill $$SERVER_PID && \
	echo "Health check passed!"

# Docker build for testing
docker-build:
	docker build -t resource-server-test .

# Docker compose for testing environment
docker-test-env:
	docker-compose -f docker-compose.test.yml up -d
	@echo "Test environment started. Database available at localhost:5432"

docker-test-env-down:
	docker-compose -f docker-compose.test.yml down

# Help target
help:
	@echo "Available targets:"
	@echo "  build              - Build the application"
	@echo "  clean              - Clean build artifacts"
	@echo "  deps               - Install dependencies"
	@echo "  test-unit          - Run unit tests"
	@echo "  test-e2e           - Run E2E tests"
	@echo "  test-workflows     - Run workflow tests only"
	@echo "  test-all           - Run all tests"
	@echo "  coverage           - Generate coverage report"
	@echo "  benchmark          - Run benchmark tests"
	@echo "  lint               - Run linter"
	@echo "  fmt                - Format code"
	@echo "  vet                - Vet code"
	@echo "  security           - Run security scan"
	@echo "  ci                 - Run full CI pipeline"
	@echo "  dev-test           - Quick development test"
	@echo "  setup-test-db      - Setup test database with Docker"
	@echo "  teardown-test-db   - Teardown test database"
	@echo "  docker-test-env    - Start test environment with Docker Compose"
	@echo "  help               - Show this help"