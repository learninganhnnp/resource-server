# CLAUDE.md

This file provides guidance to Claude Code (claude.ai/code) when working with code in this repository.

## Project Overview

This is a Go-based Resource Management API server built with Clean Architecture principles. It provides a RESTful HTTP interface for managing resource files across multiple cloud storage providers (CDN, Google Cloud Storage, Cloudflare R2). The API leverages the existing `avironactive.com/resource` domain layer for resource management operations.

## Development Commands

### Build and Run
```bash
go build -o resource-server ./cmd/api
./resource-server
```

### Testing
```bash
go test ./...
go test -v ./internal/...
```

### Linting and Formatting
```bash
go fmt ./...
go vet ./...
```

### Module Management
```bash
go mod tidy
go mod download
```

## Key Components

### Resource Definitions (`core/definition.go`)
- Defines path patterns for achievements and workouts across providers
- Configures scopes (Global, App, ClientApp) and parameter definitions
- Sets up provider-specific URL patterns and metadata configurations
- Contains hardcoded provider credentials (should be moved to config)

### Domain Integration
The codebase integrates with `avironactive.com/resource` package which provides:
- ResourceManager interface for core operations
- PathDefinition system for configurable resource paths
- Multi-provider support (CDN, GCS, R2)
- Scope management and parameter resolution
- Signed URL generation and metadata management
- Multipart upload workflow support

### Provider Configuration
Currently hardcoded in `core/definition.go:newProviders()`:
- **CDN Provider**: Base URL and signing key configuration
- **GCS Provider**: Uses Application Default Credentials
- **R2 Provider**: Account ID, access keys, and secrets (should be moved to environment variables)

## Error Handling  
- Define module-specific errors using `anerror.New*Error()` patterns:
  ```go
  var (
      ErrTokenNotFound = anerror.NewNotFoundError(213, "token", "tokenNotFound", "token not found")
      ErrExpiredToken  = anerror.NewUnauthenticatedError(217, "token", "expiredToken", "the token is expired")
  )
  ```
- Function return pattern with early returns:
  ```go
  func SomeOperation(ctx context.Context, param string) (Result, error) {
      if param == "" {
          return nil, anerror.ErrInvalidArgument.With("param is required")
      }
      if maps.Contains(someMap, param) {
          return nil, anerror.ErrAlreadyExists.Withf("param %s already exists", param)
      }
      if maps.NotContains(someMap, param) {
          return nil, anerror.ErrNotFound.Withf("param %s not found", param)
      }
      result, err := someCall(ctx, param)
      if err != nil {
          return nil, err // Pass through or wrap as needed
      }
      return result, nil
  }
  ```
- Use structured logging with error field: `logger.WithField("err", err).Error("operation failed")`
- Check error types with `errors.Is(err, anerror.ErrNotFound)` method or `anerror.As(err)`

## Testing
- Unit tests: `*_test.go`, Integration: `*_integration_test.go`
- Table-driven tests with structured input/output expectations
- Test helpers in `test/` subdirectory for fixtures
- Mock interfaces in `mock/` subdirectory with clear naming

## Modernization Notes
- Replace interface{} with any type alias
- Replace type assertions with type switches where appropriate
- Use range over patterns for loops can be modernized
- Replace the m[k]=v loop with `map.Copy` 
- Use generics for type-safe operations
- Use `context.Context` for passing request-scoped values

## Important Notes
- When you implement a task in the PRD.md file, please mark the task as done by adding a checkmark at the beginning of the line like this: "- [x] Task 1"
- The API is designed to be stateless with no authentication requirements
- Provider credentials are currently hardcoded and should be externalized to configuration
- The project uses Go 1.24.2 and depends on external `avironactive.com` package
- Clean Architecture principles should be maintained when adding new features
- All provider-specific logic is abstracted through the domain layer

