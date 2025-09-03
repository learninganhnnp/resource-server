# End-to-End Testing Plan for Resource Server APIs - ✅ COMPLETED

## Overview
This document outlines a comprehensive end-to-end testing strategy for the Resource Server API. The tests will validate all API endpoints, ensure proper integration between components, and verify business logic workflows.

**Status: ✅ IMPLEMENTATION COMPLETE**

## Test Infrastructure - ✅ COMPLETED

### Prerequisites - ✅ COMPLETED
- [x] Go 1.24.2+
- [x] Docker and Docker Compose (for testcontainers)
- [x] PostgreSQL test database
- [x] Mock storage providers (CDN, GCS, R2)

### Test Directory Structure - ✅ COMPLETED
```
test/
├── e2e/
│   ├── suite_test.go              # Main test suite setup ✅
│   ├── health_test.go             # Health check endpoint tests ✅
│   ├── resource_definition_test.go # Resource definition tests ✅
│   ├── provider_test.go          # Provider endpoint tests ✅
│   ├── file_operations_test.go   # File operation tests ✅
│   ├── multipart_test.go         # Multipart upload tests ✅
│   ├── achievement_test.go       # Achievement tests ✅
│   └── workflows_test.go         # Complex workflow tests ✅
├── fixtures/
│   ├── achievements.json         # Test achievement data ✅
│   ├── files/                    # Test files for upload ✅
│   └── metadata.json             # Test metadata ✅
└── helpers/
    ├── client.go                 # API client helpers ✅
    ├── database.go               # Database setup/teardown ✅
    ├── assertions.go             # Custom test assertions ✅
    └── mocks.go                  # Mock storage providers ✅
```

## API Endpoints Coverage

### 1. Health Check API

#### GET /api/v1/health
**Test Cases:**
| Test Case | Description | Expected Result |
|-----------|-------------|-----------------|
| HC-001 | Successful health check | Status 200, success: true |
| HC-002 | Verify response format | Contains success, data, message, error fields |
| HC-003 | Check content-type header | application/json |
| HC-004 | Verify server readiness | Database connection active |

**Sample Request/Response:**
```json
// Request
GET /api/v1/health

// Response (200 OK)
{
  "success": true,
  "data": {
    "status": "healthy",
    "timestamp": "2025-01-02T10:00:00Z"
  },
  "message": "Service is healthy",
  "error": null
}
```

### 2. Resource Definition APIs

#### GET /api/v1/resources/definitions
**Test Cases:**
| Test Case | Description | Expected Result |
|-----------|-------------|-----------------|
| RD-001 | List all definitions | Returns achievement and workout definitions |
| RD-002 | Verify definition structure | Contains name, displayName, description, parameters |
| RD-003 | Check allowed scopes | Global, App, ClientApp scopes present |
| RD-004 | Validate providers list | cdn, gcs, r2 providers listed |

**Sample Response:**
```json
{
  "success": true,
  "data": [
    {
      "name": "achievement",
      "displayName": "Achievement",
      "description": "Achievement resource definition",
      "allowedScopes": ["G", "A", "CA"],
      "parameters": [
        {
          "name": "achievementId",
          "rules": ["uuid"],
          "description": "Achievement unique identifier"
        }
      ],
      "providers": ["cdn", "gcs", "r2"]
    }
  ]
}
```

#### GET /api/v1/resources/definitions/:name
**Test Cases:**
| Test Case | Description | Expected Result |
|-----------|-------------|-----------------|
| RD-005 | Get existing definition | Returns specific definition details |
| RD-006 | Get non-existent definition | Returns 404 error |
| RD-007 | Invalid definition name format | Returns 400 validation error |
| RD-008 | Case sensitivity check | Verify exact name matching |

### 3. Provider APIs

#### GET /api/v1/resources/providers
**Test Cases:**
| Test Case | Description | Expected Result |
|-----------|-------------|-----------------|
| PR-001 | List all providers | Returns cdn, gcs, r2 providers |
| PR-002 | Verify capabilities structure | Contains all capability fields |
| PR-003 | Check multipart support | Verify multipart capabilities |
| PR-004 | Validate checksum algorithms | Supported algorithms listed |

**Sample Response:**
```json
{
  "success": true,
  "data": [
    {
      "name": "r2",
      "capabilities": {
        "supportsRead": true,
        "supportsWrite": true,
        "supportsDelete": true,
        "supportsListing": true,
        "supportsMetadata": true,
        "supportsMultipart": true,
        "supportsSignedUrls": true,
        "multipart": {
          "minPartSize": 5242880,
          "maxPartSize": 5368709120,
          "maxParts": 10000
        }
      }
    }
  ]
}
```

#### GET /api/v1/resources/providers/:name
**Test Cases:**
| Test Case | Description | Expected Result |
|-----------|-------------|-----------------|
| PR-005 | Get existing provider | Returns specific provider details |
| PR-006 | Get non-existent provider | Returns 404 error |
| PR-007 | Invalid provider name | Returns 400 validation error |

### 4. File Operations APIs

#### GET /api/v1/resources/:provider/:definition
**Test Cases:**
| Test Case | Description | Expected Result |
|-----------|-------------|-----------------|
| FO-001 | List files without filters | Returns all files |
| FO-002 | List with max_keys limit | Returns limited results |
| FO-003 | List with continuation token | Returns next page |
| FO-004 | List with prefix filter | Returns filtered files |
| FO-005 | Invalid provider | Returns 400 error |
| FO-006 | Invalid definition | Returns 400 error |
| FO-007 | Empty bucket | Returns empty list |

**Query Parameters:**
- `max_keys` (1-1000, default: 1000)
- `continuation_token` (for pagination)
- `prefix` (filter by prefix)

#### POST /api/v1/resources/:provider/:definition/upload
**Test Cases:**
| Test Case | Description | Expected Result |
|-----------|-------------|-----------------|
| FO-008 | Generate upload URL with minimal params | Returns signed URL |
| FO-009 | Upload with scope Global | URL with global scope |
| FO-010 | Upload with scope App | URL with app scope |
| FO-011 | Upload with scope ClientApp | URL with client app scope |
| FO-012 | Upload with metadata | Metadata headers included |
| FO-013 | Upload with custom expiry | Correct expiry time |
| FO-014 | Missing required parameters | Returns 400 error |
| FO-015 | Invalid scope value | Returns validation error |

**Request Body:**
```json
{
  "parameters": {
    "achievementId": "123e4567-e89b-12d3-a456-426614174000"
  },
  "scope": "G",
  "scopeValue": 0,
  "expiry": "1h",
  "metadata": {
    "contentType": "image/png",
    "cacheControl": "max-age=3600"
  }
}
```

#### POST /api/v1/resources/:provider/*/download
**Test Cases:**
| Test Case | Description | Expected Result |
|-----------|-------------|-----------------|
| FO-016 | Generate download URL for existing file | Returns signed URL |
| FO-017 | Download non-existent file | Returns 404 error |
| FO-018 | Download with custom expiry | Correct expiry time |
| FO-019 | Download with response headers | Headers included in URL |
| FO-020 | Invalid file path | Returns 400 error |

#### GET /api/v1/resources/:provider/*/metadata
**Test Cases:**
| Test Case | Description | Expected Result |
|-----------|-------------|-----------------|
| FO-021 | Get metadata for existing file | Returns file metadata |
| FO-022 | Get metadata for non-existent file | Returns 404 error |
| FO-023 | Verify all metadata fields | All fields present |
| FO-024 | Check custom metadata | Custom headers returned |

#### PUT /api/v1/resources/:provider/*/metadata
**Test Cases:**
| Test Case | Description | Expected Result |
|-----------|-------------|-----------------|
| FO-025 | Update content type | Content type updated |
| FO-026 | Update cache control | Cache control updated |
| FO-027 | Update custom metadata | Custom headers updated |
| FO-028 | Update non-existent file | Returns 404 error |
| FO-029 | Invalid metadata values | Returns validation error |

#### DELETE /api/v1/resources/:provider/*
**Test Cases:**
| Test Case | Description | Expected Result |
|-----------|-------------|-----------------|
| FO-030 | Delete existing file | File deleted successfully |
| FO-031 | Delete non-existent file | Returns 404 error |
| FO-032 | Delete with invalid path | Returns 400 error |
| FO-033 | Verify file removal | File not listed after delete |

### 5. Multipart Upload APIs

#### POST /api/v1/resources/multipart/init
**Test Cases:**
| Test Case | Description | Expected Result |
|-----------|-------------|-----------------|
| MP-001 | Initialize multipart upload | Returns upload ID and constraints |
| MP-002 | Init with minimal parameters | Success with defaults |
| MP-003 | Init with metadata | Metadata stored |
| MP-004 | Invalid provider | Returns 400 error |
| MP-005 | Invalid definition | Returns 400 error |
| MP-006 | Missing required fields | Returns validation error |

**Request Body:**
```json
{
  "definitionName": "achievement",
  "provider": "r2",
  "scope": "G",
  "scopeValue": 0,
  "paramResolver": {
    "achievementId": "123e4567-e89b-12d3-a456-426614174000"
  },
  "metadata": {
    "contentType": "image/png",
    "contentEncoding": "gzip"
  }
}
```

#### POST /api/v1/resources/multipart/urls
**Test Cases:**
| Test Case | Description | Expected Result |
|-----------|-------------|-----------------|
| MP-007 | Get part URLs for valid upload | Returns signed URLs for parts |
| MP-008 | Get URLs with checksums | URLs include checksum headers |
| MP-009 | Request too many parts | Returns validation error |
| MP-010 | Invalid upload ID | Returns 404 error |
| MP-011 | Get complete and abort URLs | Both URLs returned |

**Request Body:**
```json
{
  "path": "achievements/icons/achievement123.png",
  "uploadId": "upload-123",
  "provider": "r2",
  "urlOptions": [
    {
      "partNumber": 1,
      "checksum": {
        "algorithm": "SHA256",
        "value": "abc123..."
      }
    }
  ]
}
```

### 6. Achievement APIs

#### GET /api/v1/achievements/
**Test Cases:**
| Test Case | Description | Expected Result |
|-----------|-------------|-----------------|
| AC-001 | List all achievements | Returns paginated list |
| AC-002 | List with pagination | Correct page results |
| AC-003 | List only active achievements | only_active=true filter |
| AC-004 | List with custom page size | Respects pageSize |
| AC-005 | Invalid page number | Defaults to page 1 |
| AC-006 | Page size exceeds limit | Capped at 100 |

**Query Parameters:**
- `page` (default: 1)
- `pageSize` (1-100, default: 20)
- `only_active` (default: true)

#### POST /api/v1/achievements/
**Test Cases:**
| Test Case | Description | Expected Result |
|-----------|-------------|-----------------|
| AC-007 | Create achievement without icon | Achievement created, upload URL provided |
| AC-008 | Create with all fields | All fields stored |
| AC-009 | Create with icon format | Upload URL for icon |
| AC-010 | Missing required name | Returns validation error |
| AC-011 | Invalid points range | Returns validation error |
| AC-012 | Invalid icon format | Returns validation error |
| AC-013 | Duplicate achievement name | Returns conflict error |

**Request Body:**
```json
{
  "name": "First Achievement",
  "description": "Complete your first workout",
  "category": "fitness",
  "points": 100,
  "iconFormat": "png",
  "provider": "r2"
}
```

#### GET /api/v1/achievements/:id
**Test Cases:**
| Test Case | Description | Expected Result |
|-----------|-------------|-----------------|
| AC-014 | Get existing achievement | Returns achievement details |
| AC-015 | Get non-existent achievement | Returns 404 error |
| AC-016 | Invalid UUID format | Returns 400 error |
| AC-017 | Verify all fields returned | All fields present |

#### PUT /api/v1/achievements/:id/icon
**Test Cases:**
| Test Case | Description | Expected Result |
|-----------|-------------|-----------------|
| AC-018 | Update icon for existing achievement | Returns new upload URL |
| AC-019 | Update non-existent achievement | Returns 404 error |
| AC-020 | Invalid format | Returns validation error |
| AC-021 | Invalid provider | Returns validation error |

**Request Body:**
```json
{
  "format": "webp",
  "provider": "cdn"
}
```

#### POST /api/v1/achievements/uploads/:id/confirm
**Test Cases:**
| Test Case | Description | Expected Result |
|-----------|-------------|-----------------|
| AC-022 | Confirm successful upload | Upload marked complete |
| AC-023 | Confirm failed upload | Upload marked failed |
| AC-024 | Confirm with metadata | Metadata stored |
| AC-025 | Confirm non-existent upload | Returns 404 error |
| AC-026 | Invalid upload ID | Returns 400 error |
| AC-027 | Confirm with ETags (multipart) | ETags processed |

**Request Body:**
```json
{
  "success": true,
  "file_size": 1024000,
  "content_type": "image/png",
  "etags": [
    {"part": 1, "etag": "abc123", "size": 5242880},
    {"part": 2, "etag": "def456", "size": 1024000}
  ],
  "metadata": {
    "contentType": "image/png",
    "cacheControl": "max-age=3600"
  },
  "verify_exists": true
}
```

#### POST /api/v1/achievements/uploads/:id/multipart
**Test Cases:**
| Test Case | Description | Expected Result |
|-----------|-------------|-----------------|
| AC-028 | Get multipart URLs for valid upload | Returns part URLs |
| AC-029 | Request invalid part count | Returns validation error |
| AC-030 | Non-existent upload ID | Returns 404 error |
| AC-031 | Part count exceeds limit | Returns validation error |

## Complex Workflow Tests

### Workflow 1: Complete Achievement Creation with Icon Upload
1. Create achievement with icon format specified
2. Receive upload URL
3. Simulate file upload to URL
4. Confirm upload completion
5. Verify achievement has icon URL
6. Download icon via generated URL

### Workflow 2: Large File Multipart Upload
1. Initialize multipart upload for large icon (>5MB)
2. Get part URLs for multiple parts
3. Simulate uploading each part
4. Complete multipart upload with ETags
5. Confirm upload completion
6. Verify file exists and metadata correct

### Workflow 3: File Lifecycle Management
1. Generate upload URL
2. Upload file
3. List files to verify existence
4. Get file metadata
5. Update file metadata
6. Generate download URL
7. Delete file
8. Verify file no longer exists

### Workflow 4: Achievement Management Flow
1. Create multiple achievements
2. List achievements with pagination
3. Update achievement icons
4. Filter active/inactive achievements
5. Verify category filtering

### Workflow 5: Cross-Provider Operations
1. Upload to CDN provider
2. Upload to GCS provider
3. Upload to R2 provider
4. List files from each provider
5. Verify provider-specific capabilities

## Error Handling Tests

### Validation Errors
- Invalid UUIDs
- Missing required fields
- Invalid enum values
- String length violations
- Numeric range violations

### Business Logic Errors
- Non-existent resources
- Duplicate resources
- Invalid state transitions
- Permission violations

### Infrastructure Errors
- Database connection failures
- Storage provider timeouts
- Network failures
- Rate limiting

## Performance Tests

### Load Testing Scenarios
1. **Concurrent Uploads**: 100 simultaneous upload URL generations
2. **File Listing**: List 10,000 files with pagination
3. **Multipart Upload**: Upload 100MB file in 20 parts
4. **Achievement Query**: Query 1000 achievements with filters

### Performance Metrics
- Response time (p50, p95, p99)
- Throughput (requests/second)
- Error rate
- Database connection pool usage
- Memory consumption

## Test Data Management

### Fixtures
```go
// test/fixtures/achievements.go
var TestAchievements = []Achievement{
    {
        Name: "First Steps",
        Description: "Complete your first workout",
        Category: "beginner",
        Points: 10,
    },
    {
        Name: "Marathon Runner",
        Description: "Complete 42km in total",
        Category: "endurance",
        Points: 500,
    },
}
```

### Database Seeding
- Create test database with schema
- Seed initial achievements
- Create test user contexts
- Setup provider configurations

### Cleanup Strategy
- Truncate tables after each test suite
- Remove uploaded test files
- Clear cache and temporary data
- Reset mock provider state

## Implementation Guidelines

### Test Structure
```go
package e2e

import (
    "testing"
    "github.com/stretchr/testify/suite"
)

type AchievementTestSuite struct {
    suite.Suite
    server *TestServer
    db     *TestDatabase
}

func (s *AchievementTestSuite) SetupSuite() {
    // Initialize test server and database
}

func (s *AchievementTestSuite) TearDownSuite() {
    // Cleanup resources
}

func (s *AchievementTestSuite) TestCreateAchievement() {
    // Test implementation
}

func TestAchievementSuite(t *testing.T) {
    suite.Run(t, new(AchievementTestSuite))
}
```

### Assertion Helpers
```go
// helpers/assertions.go
func AssertSuccessResponse(t *testing.T, resp *http.Response) {
    var body SuccessResponse
    json.NewDecoder(resp.Body).Decode(&body)
    assert.True(t, body.Success)
    assert.Nil(t, body.Error)
}

func AssertErrorResponse(t *testing.T, resp *http.Response, code string) {
    var body ErrorResponse
    json.NewDecoder(resp.Body).Decode(&body)
    assert.False(t, body.Success)
    assert.Equal(t, code, body.Error.Code)
}
```

### Mock Providers
```go
// helpers/mocks.go
type MockStorageProvider struct {
    mock.Mock
}

func (m *MockStorageProvider) GenerateUploadURL(ctx context.Context, path string, opts ...Option) (*SignedURL, error) {
    args := m.Called(ctx, path, opts)
    return args.Get(0).(*SignedURL), args.Error(1)
}
```

## CI/CD Integration

### GitHub Actions Workflow
```yaml
name: E2E Tests
on: [push, pull_request]
jobs:
  test:
    runs-on: ubuntu-latest
    services:
      postgres:
        image: postgres:15
        env:
          POSTGRES_PASSWORD: testpass
        options: >-
          --health-cmd pg_isready
          --health-interval 10s
          --health-timeout 5s
          --health-retries 5
    steps:
      - uses: actions/checkout@v3
      - uses: actions/setup-go@v4
        with:
          go-version: '1.24.2'
      - name: Run E2E Tests
        run: |
          go test -v ./test/e2e/... -tags=e2e
```

### Test Execution Commands
```bash
# Run all E2E tests
go test -v ./test/e2e/... -tags=e2e

# Run specific test suite
go test -v ./test/e2e/achievement_test.go -tags=e2e

# Run with coverage
go test -v -coverprofile=coverage.out ./test/e2e/... -tags=e2e

# Run with race detection
go test -v -race ./test/e2e/... -tags=e2e

# Run performance tests
go test -v ./test/e2e/... -tags=e2e,performance -bench=.
```

## Success Criteria

### Coverage Requirements
- API endpoint coverage: 100%
- Business logic coverage: >80%
- Error handling coverage: >90%
- Integration workflow coverage: 100%

### Quality Metrics
- All tests pass consistently
- No flaky tests
- Test execution time <5 minutes
- Clear test documentation
- Maintainable test code

### Deliverables
1. Complete E2E test suite implementation
2. Test documentation and reports
3. CI/CD pipeline configuration
4. Performance test baselines
5. Test data management scripts

## Timeline

| Phase | Duration | Deliverables |
|-------|----------|--------------|
| Infrastructure Setup | 2 days | Test framework, database, mocks |
| API Tests Implementation | 5 days | All endpoint tests |
| Workflow Tests | 2 days | Complex scenario tests |
| Performance Tests | 2 days | Load and stress tests |
| CI/CD Integration | 1 day | Automated pipeline |
| Documentation | 1 day | Complete test documentation |

## Maintenance

### Regular Tasks
- Update tests for API changes
- Review and update test data
- Monitor test execution times
- Investigate flaky tests
- Update mock providers
- Performance baseline updates

### Test Review Checklist
- [ ] All new endpoints have tests
- [ ] Error cases are covered
- [ ] Performance impact assessed
- [ ] Documentation updated
- [ ] CI pipeline passing
- [ ] No hardcoded values
- [ ] Proper cleanup implemented