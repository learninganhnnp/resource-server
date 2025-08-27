# Resource Management API - Product Requirements Document

## Executive Summary

The Resource Management API is a RESTful HTTP service built with Go Fiber that provides a unified interface for managing resource files across multiple cloud storage providers. The API leverages the existing `avironactive.com/resource` domain layer and follows Clean Architecture principles to ensure maintainability, testability, and separation of concerns.

## Problem Statement

Currently, resource management operations (file uploads, downloads, metadata management) across multiple cloud providers require direct integration with the domain layer. There's a need for a standardized HTTP API that:

1. Provides a unified interface for resource operations
2. Abstracts provider-specific implementations 
3. Supports multiple cloud storage providers (Cloudflare R2, Google Cloud Storage, CDN)
4. Handles complex multipart upload workflows
5. Manages resource path definitions and provider configurations

## Goals and Objectives

### Primary Goals
- **Unified API Interface**: Single HTTP API for all resource operations
- **Multi-Provider Support**: Seamless integration with R2, GCS, and CDN providers
- **Clean Architecture**: Maintainable codebase with clear layer separation
- **Stateless Operations**: No authentication requirements, fully stateless design
- **Domain Reuse**: Leverage existing `avironactive.com/resource` package without duplication

### Success Metrics
- API response time < 500ms for standard operations
- Support for files up to 5GB via multipart upload
- 99.9% uptime for core operations
- Zero breaking changes to existing domain layer

## User Stories

### As a Client Application Developer
- I want to list available resource definitions so I can understand what resources are available
- I want to get specific resource definitions so I can understand their parameters and constraints
- I want to generate signed URLs for file uploads so I can upload files directly to storage providers
- I want to generate signed URLs for file downloads so users can access files securely
- I want to initiate multipart uploads for large files so I can handle files larger than memory limits
- I want to list files in a resource path so I can display available resources to users
- I want to get file metadata so I can display file information without downloading the file
- I want to update file metadata so I can modify cache settings and access controls
- I want to delete files so I can manage storage usage

### As a System Administrator
- I want to manage provider configurations so I can control which storage providers are used
- I want to monitor API usage so I can ensure system performance
- I want to configure resource definitions so I can control how resources are organized

## Functional Requirements

### 1. Resource Definition Management

#### 1.1 List Resource Definitions
- **Endpoint**: `GET /api/v1/resources/definitions`
- **Description**: Retrieve all available resource path definitions
- **Response**: Array of definition objects with name, display name, description, allowed scopes, and parameters

#### 1.2 Get Resource Definition
- **Endpoint**: `GET /api/v1/resources/definitions/:name`
- **Description**: Retrieve a specific resource definition by name
- **Parameters**: 
  - `name` (path): Resource definition name
- **Response**: Single definition object with full details including patterns and metadata

### 2. Provider Management

#### 2.1 List Providers
- **Endpoint**: `GET /api/v1/resources/providers`
- **Description**: Retrieve all configured storage providers
- **Response**: Array of provider objects with name and capabilities 

#### 2.2 Get Provider Details
- **Endpoint**: `GET /api/v1/resources/providers/:name`
- **Description**: Retrieve detailed information about a specific provider
- **Parameters**:
  - `name` (path): Provider name (cdn, gcs, r2)
- **Response**: Provider object with capabilities, constraints, and configuration details

### 3. File Operations

#### 3.1 List Files
- **Endpoint**: `GET /api/v1/resources/:provider/:definition`
- **Description**: List files in a resource path with pagination
- **Parameters**:
  - `provider` (path): Storage provider name
  - `definition` (path): Resource definition name
  - Query parameters for pagination and filtering
- **Response**: Paginated list of files with metadata

#### 3.2 Generate Upload URL
- **Endpoint**: `POST /api/v1/resources/:provider/:definition/upload`
- **Description**: Generate signed URL for file upload
- **Parameters**:
  - `provider` (path): Storage provider name
  - `definition` (path): Resource definition name
  - Request body with file parameters and metadata
- **Response**: Signed URL with expiry and required headers

#### 3.3 Generate Multipart Upload URLs
- **Endpoint**: `POST /api/v1/resources/:provider/:definition/upload/multipart`
- **Description**: Generate URLs for multipart upload workflow
- **Parameters**:
  - `provider` (path): Storage provider name
  - `definition` (path): Resource definition name
  - Request body with file parameters and part configuration
- **Response**: Multipart upload URLs and constraints

#### 3.4 Generate Download URL
- **Endpoint**: `POST /api/v1/resources/:provider/*/download`
- **Description**: Generate signed URL for file download
- **Parameters**:
  - `provider` (path): Storage provider name
  - `*` (path): File path (wildcard to support nested paths)
  - Request body with download parameters
- **Response**: Signed URL with expiry and response headers

### 4. Metadata Management

#### 4.1 Get File Metadata
- **Endpoint**: `GET /api/v1/resources/:provider/*/metadata`
- **Description**: Retrieve metadata for a specific file
- **Parameters**:
  - `provider` (path): Storage provider name
  - `*` (path): File path
- **Response**: File metadata including size, content type, cache settings, and custom metadata

#### 4.2 Update File Metadata
- **Endpoint**: `PUT /api/v1/resources/:provider/*/metadata`
- **Description**: Update metadata for an existing file
- **Parameters**:
  - `provider` (path): Storage provider name
  - `*` (path): File path
  - Request body with metadata updates
- **Response**: Updated metadata object

### 5. File Deletion

#### 5.1 Delete File
- **Endpoint**: `DELETE /api/v1/resources/:provider/*`
- **Description**: Delete a file from storage
- **Parameters**:
  - `provider` (path): Storage provider name
  - `*` (path): File path
- **Response**: Success confirmation

## Technical Requirements

### 1. Architecture

#### 1.1 Clean Architecture Layers
```
┌─────────────────────────────────────┐
│     HTTP Layer (Fiber Handlers)     │  ← REST API endpoints
├─────────────────────────────────────┤
│      Application Layer (Use Cases)  │  ← Business logic orchestration
├─────────────────────────────────────┤
│    Domain Layer (avironactive.com)  │  ← Existing resource package (reused)
├─────────────────────────────────────┤
│   Infrastructure Layer (Config)     │  ← Configuration and external services
└─────────────────────────────────────┘
```

#### 1.2 Directory Structure
```
├── cmd/
│   └── api/                 # Application entry point
├── internal/
│   ├── app/
│   │   ├── usecases/        # Use case implementations
│   │   └── dto/             # Data transfer objects
│   ├── infrastructure/
│   │   ├── config/          # Configuration management
│   │   └── http/            # HTTP server setup
│   └── interfaces/
│       └── http/
│           ├── handlers/    # HTTP route handlers
│           └── middleware/  # HTTP middleware
├── pkg/
│   └── errors/              # Custom error types
└── core/                    # Resource definitions (existing)
```

### 2. Technology Stack

#### 2.1 Core Framework
- **Web Framework**: Go Fiber v2 (Express-inspired, high performance)
- **Language**: Go 1.24.2
- **Domain Layer**: Existing `avironactive.com/resource` package

#### 2.2 Configuration
- **Format**: YAML with environment variable overrides
- **Library**: Built-in YAML parsing with validation
- **Hot Reload**: Configuration reload without restart

#### 2.3 Dependencies
- **HTTP Router**: Fiber v2 with middleware support
- **JSON Handling**: Built-in encoding/json
- **Validation**: Custom validation for DTOs
- **Logging**: Integration with existing zap logger
- **CORS**: Fiber CORS middleware

### 3. API Design Standards

#### 3.1 RESTful Conventions
- **HTTP Methods**: GET, POST, PUT, DELETE following REST semantics
- **Status Codes**: Standard HTTP status codes (200, 201, 400, 404, 500)
- **Content Type**: application/json for all requests/responses
- **Error Format**: Consistent error response structure
- **Formatting**: Use of camelCase for JSON fields

#### 3.2 Response Format
```json
{
  "success": true,
  "data": {...},
  "message": "Optional message",
  "error": null
}
```

#### 3.3 Error Format
```json
{
  "success": false,
  "data": null,
  "message": "Error description",
  "error": {
    "code": "ERROR_CODE",
    "details": "Detailed error information"
  }
}
```

### 4. Security Requirements

#### 4.1 No Authentication
- API operates without authentication as specified
- All security handled at infrastructure level
- Rate limiting can be implemented at reverse proxy

#### 4.2 Input Validation
- All request parameters validated
- File path sanitization to prevent directory traversal
- Provider name validation against allowed providers

#### 4.3 CORS Support
- Configurable CORS policies
- Support for cross-origin requests
- Preflight request handling

### 5. Performance Requirements

#### 5.1 Response Time
- Standard operations: < 200ms
- File listing: < 500ms
- URL generation: < 100ms
- Metadata operations: < 300ms

#### 5.2 Throughput
- Handle 1000+ concurrent requests
- Support large file operations via multipart upload
- Efficient memory usage for file operations

#### 5.3 Scalability
- Stateless design for horizontal scaling
- Resource manager connection pooling
- Configurable timeout and retry policies

## API Endpoints Specification

### Resource Definitions

```http
GET /api/v1/resources/definitions
Response: Array of PathDefinition objects

GET /api/v1/resources/definitions/:name
Parameters:
  - name: string (path parameter)
Response: Single PathDefinition object
```

### Provider Management

```http
GET /api/v1/resources/providers
Response: Array of Provider objects with capabilities

GET /api/v1/resources/providers/:name
Parameters:
  - name: string (cdn|gcs|r2)
Response: Provider object with detailed capabilities
```

### File Operations

```http
GET /api/v1/resources/:provider/:definition
Parameters:
  - provider: string (path)
  - definition: string (path)
  - max_keys: integer (query, optional, default: 1000)
  - continuation_token: string (query, optional)
  - prefix: string (query, optional)
Response: Paginated file listing

POST /api/v1/resources/:provider/:definition/upload
Parameters:
  - provider: string (path)
  - definition: string (path)
Request Body: UploadRequest
Response: SignedURL object

POST /api/v1/resources/:provider/:definition/upload/multipart
Parameters:
  - provider: string (path)  
  - definition: string (path)
Request Body: MultipartUploadRequest
Response: MultipartUpload object

POST /api/v1/resources/:provider/*/download
Parameters:
  - provider: string (path)
  - *: string (file path)
Request Body: DownloadRequest
Response: SignedURL object
```

### Metadata Operations

```http
GET /api/v1/resources/:provider/*/metadata
Parameters:
  - provider: string (path)
  - *: string (file path)
Response: FileMetadata object

PUT /api/v1/resources/:provider/*/metadata  
Parameters:
  - provider: string (path)
  - *: string (file path)
Request Body: MetadataUpdate
Response: FileMetadata object
```

### File Deletion

```http
DELETE /api/v1/resources/:provider/*
Parameters:
  - provider: string (path)
  - *: string (file path)
Response: Success confirmation
```

## Data Models

### Request DTOs

#### UploadRequest
```json
{
  "parameters": {
    "param_name": "value"
  },
  "scope": "G|A|CA",
  "scope_value": 123,
  "expiry": "24h",
  "metadata": {
    "content_type": "image/png",
    "cache_control": "max-age=3600",
    "custom_headers": {}
  }
}
```

#### MultipartUploadRequest
```json
{
  "parameters": {
    "param_name": "value"  
  },
  "scope": "G|A|CA",
  "scope_value": 123,
  "file_size": 1073741824,
  "part_count": 200,
  "metadata": {}
}
```

#### DownloadRequest
```json
{
  "expiry": "1h",
  "response_headers": {
    "content_disposition": "attachment; filename=file.pdf"
  }
}
```

### Response DTOs

#### PathDefinition
```json
{
  "name": "achievements",
  "display_name": "Achievement Resources", 
  "description": "Achievement icons and metadata",
  "allowed_scopes": ["A", "G"],
  "parameters": [
    {
      "name": "app",
      "required": false,
      "description": "Application name",
      "default_value": ""
    }
  ],
  "providers": ["cdn", "r2"]
}
```

#### SignedURL
```json
{
  "url": "https://storage.example.com/signed-url",
  "method": "PUT",
  "headers": {
    "Content-Type": "image/png"
  },
  "expires_at": "2024-01-01T12:00:00Z"
}
```

#### FileMetadata
```json
{
  "key": "path/to/file.png",
  "size": 1024,
  "content_type": "image/png",
  "etag": "abc123",
  "last_modified": "2024-01-01T12:00:00Z",
  "cache_control": "max-age=3600",
  "metadata": {}
}
```

## Configuration

### Application Configuration (YAML)

```yaml
server:
  port: 8080
  host: "0.0.0.0"
  read_timeout: "30s"
  write_timeout: "30s"

cors:
  allowed_origins: ["*"]
  allowed_methods: ["GET", "POST", "PUT", "DELETE", "OPTIONS"]
  allowed_headers: ["Content-Type", "Authorization"]

providers:
  cdn:
    base_url: "https://cdn-dev.avironactive.net/assets"
    signing_key: "${CDN_SIGNING_KEY}"
    expiry: "24h"
  
  gcs:
    expiry: "24h"
    # GCS credentials via service account key or ADC
  
  r2:
    account_id: "${R2_ACCOUNT_ID}"
    access_key_id: "${R2_ACCESS_KEY_ID}"  
    secret_key: "${R2_SECRET_KEY}"
    expiry: "24h"

logging:
  level: "info"
  format: "json"
```

## Implementation Phases

### Phase 1: Foundation (Week 1)
- [x] Project structure setup
- [x] Configuration management
- [x] Basic Fiber application
- [x] Error handling middleware
- [x] Health check endpoint

### Phase 2: Core API (Week 2)  
- [x] Resource definitions endpoints
- [x] Provider management endpoints
- [x] Basic file operations
- [x] Request/response DTOs

### Phase 3: Advanced Features (Week 3)
- [x] Multipart upload workflow
- [x] Metadata management
- [x] File deletion
- [x] Input validation

### Phase 4: Polish & Testing (Week 4)
- [ ] Integration tests
- [ ] Performance optimization
- [ ] Documentation
- [ ] Deployment configuration

## Success Criteria

### Functional Success
- [ ] All API endpoints implemented and working
- [ ] Integration with existing domain layer without modifications
- [ ] Support for all three providers (CDN, GCS, R2)
- [ ] Multipart upload workflow functional
- [ ] Proper error handling and validation

### Technical Success
- [ ] Clean architecture maintained
- [ ] Response times meet performance requirements  
- [ ] Memory usage optimized for large files
- [ ] Configuration system flexible and environment-aware
- [ ] Comprehensive test coverage

### Operational Success  
- [ ] API documentation complete
- [ ] Deployment pipeline ready
- [ ] Monitoring and logging configured
- [ ] Error tracking implemented
- [ ] Performance metrics available

## Risks and Mitigation

### Technical Risks
- **Domain Layer Changes**: Risk that existing domain layer needs modifications
  - *Mitigation*: Thorough analysis completed, use adapter pattern if needed
- **Provider Configuration**: Complex provider setup might cause issues
  - *Mitigation*: Comprehensive configuration validation and clear error messages
- **Memory Usage**: Large file operations could cause memory issues  
  - *Mitigation*: Stream-based processing and multipart uploads

### Operational Risks
- **Performance**: API might not meet performance requirements
  - *Mitigation*: Early performance testing and optimization
- **Scalability**: Stateless design might have hidden state dependencies
  - *Mitigation*: Thorough review of domain layer for any stateful components

## Appendix

### Domain Layer Analysis Summary

The existing `avironactive.com/resource` package provides:

- **ResourceManager Interface**: Core operations for file management
- **PathDefinition System**: Configurable resource path definitions with parameters
- **Multi-Provider Support**: CDN, GCS, R2 providers with capabilities
- **Scope Management**: Global, App, ClientApp scopes with context-aware selection
- **URL Generation**: Signed URL generation for read/write operations
- **Metadata Management**: Storage metadata configuration and management
- **Multipart Support**: Full multipart upload workflow
- **Template Resolution**: Parameter injection and path resolution

This comprehensive domain layer requires no modifications and can be directly integrated into the Clean Architecture approach.


Open questions:
- should we use pattern factory method for convert object to model?
like 
```type ModelResponse struct {
	// fields...
}
func NewModelResponseFromDTO(dto DTO) *ModelResponse {
   return &ModelResponse{
	  // mapping fields...
   }
}

```

``` type ModelRequest struct {
	// fields...
}

func (r *ModelRequest) To() *Options {
	return &Options{
		// mapping fields...
	}
}