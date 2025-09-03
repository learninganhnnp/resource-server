# Resource Package Implementation Plan (Updated)

## Executive Summary
Based on comprehensive analysis of the existing codebase and the proposed `update-resource-flow.md`, the resource package already demonstrates strong DDD principles with proper bounded contexts. The implementation plan focuses on enhancing the upload workflow to align with the database schema design and improving service separation for better maintainability.

## Bounded Context Definition

### Resource Management Bounded Context
**Core Responsibility:** File and resource lifecycle management including upload, storage, and delivery operations.

**Bounded Context Boundaries:**
- **Includes:** Upload workflows, path resolution, provider abstraction, storage metadata
- **Excludes:** File content processing, user authentication, business-specific workflows, notifications, resource identification

**Ubiquitous Language:**
- **Resource:** Any file or digital asset managed by the system (images, documents, videos)
- **Upload:** Process of transferring files to storage providers
- **Provider:** Storage backend (S3, GCS, R2, CDN)
- **Path Definition:** Template-based rules for generating storage paths
- **Scope:** Context for path resolution (App, ClientApp, Global)
- **Multipart Upload:** Chunked upload process for large files
- **Storage Key:** Unique identifier for stored objects
- **Delivery Path:** URL path for accessing resources
- **Signed URL:** Time-limited, secure access URL

**Context Responsibilities:**
- Upload orchestration and state management
- Storage provider abstraction and selection
- Path template resolution with parameters
- URL generation for upload/download operations
- Resource metadata management

**External Integrations:**
- Storage providers (AWS S3, Google Cloud Storage, Cloudflare R2)
- CDN services for content delivery
- Database for upload state persistence
- Logging and metrics systems

### Domain Interaction Map
```
┌─────────────────────────────────────────────────────────────┐
│                Resource Management Bounded Context           │
│                                                             │
│  ┌─────────────────────────┐                                │
│  │   ResourceManager       │ (Application Service)           │
│  │   (Orchestration)       │                                │
│  └─────────┬───────────────┘                                │
│            │                                                │
│            ▼                                                │
│  ┌─────────────────────────┐                                │
│  │    Upload Domain        │ (Business Logic)                │
│  │ ┌─────────────────────┐ │                                │
│  │ │ Simple Upload Svc   │ │                                │
│  │ │ Multipart Upload    │ │                                │
│  │ │ Upload Facade       │ │                                │
│  │ │ Upload Aggregate    │ │                                │
│  │ └─────────────────────┘ │                                │
│  └─────────┬───────────────┘                                │
│            │                                                │
│            ▼                                                │
│  ┌─────────────────────────────────────────────────────────┐ │
│  │                SHARED DOMAINS                           │ │
│  │ ┌─────────────────┐ ┌─────────────────────────────────┐ │ │
│  │ │ Path Resolution │ │    Provider Abstraction        │ │ │
│  │ │ • Templates     │ │    • S3, GCS, R2, CDN         │ │ │
│  │ │ • Parameters    │ │    • Multipart Detection       │ │ │
│  │ │ • Scopes        │ │    • URL Generation            │ │ │
│  │ │ • Definitions   │ │    • Smart Selection           │ │ │
│  │ └─────────────────┘ └─────────────────────────────────┘ │ │
│  │ ┌─────────────────┐                                     │ │
│  │ │ Storage Metadata│                                     │ │
│  │ │ • Content-Type  │                                     │ │
│  │ │ • ACL           │                                     │ │
│  │ │ • Cache Control │                                     │ │
│  │ │ • Checksums     │                                     │ │
│  │ └─────────────────┘                                     │ │
│  └─────────────────────────────────────────────────────────┘ │
└─────────────────────────────────────────────────────────────┘

External Context Boundaries:
• Storage Providers (AWS S3, GCS, R2)
• CDN Services
• Database Persistence  
• Authentication/Authorization Context
• Notification Context
```

## Analysis of Current State

### Current Architecture Strengths

Based on the codebase analysis:

1. **Well-Structured Packages**
   - `metadata/`: Comprehensive storage metadata management
   - `provider/`: Clean provider abstractions with multiple implementations
   - `resolver/`: Sophisticated template-based path resolution
   - `upload/`: Rich domain model with Upload aggregate

2. **Implemented DDD Patterns**
   - Upload aggregate with rich behavior methods
   - Clear bounded contexts
   - Proper value objects (UploadID, PathResolutionOptions)
   - Repository pattern for persistence

3. **Areas for Enhancement**
   - Align Upload model with proposed database schema
   - Separate simple vs multipart upload services
   - Implement deterministic path generation workflow
   - Add resource_value field for CDN path tracking

### 2. Shared Domains Within Bounded Context

#### Provider Domain (Shared Infrastructure)
- **Core Responsibility:** Storage provider abstraction and management
- **Key Components:** 
  - Provider interface and implementations (S3, GCS, R2, CDN)
  - Multipart capabilities detection
  - Signed URL generation (upload/download)
  - Object metadata management
  - Provider selection and fallback logic
- **Current Location:** `resource/provider/`
- **Shared By:** Upload services, download operations, resource management

#### Path Resolution Domain (Shared Infrastructure)
- **Core Responsibility:** Path template resolution and URL generation
- **Key Components:**
  - Template pattern resolution (`{param}` replacement)
  - Parameter validation and type conversion
  - Scope-based routing (App, ClientApp, Global scopes)
  - Path definition registry and management
  - URL pattern matching and generation
- **Current Location:** `resource/resolver/`
- **Shared By:** Upload services, download operations, resource metadata operations

#### Storage Metadata Domain (Shared)
- **Core Responsibility:** Resource metadata management
- **Key Components:**
  - Content-Type determination
  - Cache control policies
  - ACL (Access Control List) management
  - Custom metadata key-value pairs
  - Checksums and validation
  - Storage class optimization
- **Shared By:** Upload operations, download operations, file management


#### Upload Domain (Business Logic)
- **Core Responsibility:** Upload workflow and state management
- **Key Components:**
  - Upload aggregate root with rich behavior
  - Status transitions and validation
  - Progress tracking (especially multipart)
  - Error handling and retry logic
  - Simple vs Multipart upload coordination
- **Current Location:** `resource/upload/`
- **Business Logic:** Contains upload-specific business rules and workflows

### Key Implementation Requirements from Flow Analysis

Based on `update-resource-flow.md`:

1. **Database Schema Alignment**
   - Add `resource_value` field for CDN path context
   - Enhance ETags structure for unified handling
   - Support both simple and multipart in single table

2. **Upload Workflow Pattern**
   - Generate deterministic paths upfront
   - Set business entity path immediately
   - Track upload status separately
   - Support retry with same path

3. **Service Separation Needs**
   - Distinct services for simple vs multipart
   - Unified facade for backward compatibility
   - Clear responsibility boundaries

## Detailed Implementation Plan

### Phase 1: Separate Upload Services

#### 1.1 Create Service Interfaces
```go
// SimpleUploadService handles direct file uploads
type SimpleUploadService interface {
    InitiateUpload(ctx context.Context, opts *UploadOptions) (*Upload, *ResolvedResource, error)
    ConfirmUpload(ctx context.Context, uploadID UploadID, confirmation *UploadConfirmation) error
    GetUploadURL(ctx context.Context, uploadID UploadID) (*provider.ObjectURL, error)
}

// MultipartUploadService handles chunked file uploads
type MultipartUploadService interface {
    InitiateUpload(ctx context.Context, opts *UploadOptions) (*Upload, *ResolvedResource, error)
    InitializeMultipart(ctx context.Context, uploadID UploadID, multipartID string, totalParts int) error
    GetPartURLs(ctx context.Context, uploadID UploadID, partCount int) (*provider.MultipartURLs, error)
    RecordPartUpload(ctx context.Context, uploadID UploadID, partNumber int, etag string, size int64) error
    CompleteUpload(ctx context.Context, uploadID UploadID, parts []PartETag) error
    ConfirmUpload(ctx context.Context, uploadID UploadID, confirmation *UploadConfirmation) error
}

// UploadServiceFacade provides unified interface
type UploadServiceFacade interface {
    InitiateUpload(ctx context.Context, opts *UploadOptions) (*Upload, *ResolvedResource, error)
    GetService(uploadType UploadType) (interface{}, error)
    // Common operations
    GetUpload(ctx context.Context, uploadID UploadID) (*Upload, error)
    ListActiveUploads(ctx context.Context, resourceType, resourceID string) ([]*Upload, error)
    AbortUpload(ctx context.Context, uploadID UploadID, reason string) error
    CleanupExpiredUploads(ctx context.Context) error
}
```

#### 1.2 Implementation Structure
```
resource/upload/
├── service_simple.go        # SimpleUploadService implementation
├── service_multipart.go     # MultipartUploadService implementation
├── service_facade.go        # UploadServiceFacade implementation
├── service_base.go          # Shared logic and helpers
└── upload_service.go        # Keep for backward compatibility (deprecated)
```

### Phase 2: Enhance ResourceManager

#### 2.1 Refactor ResourceManager Interface
```go
type ResourceManager interface {
    // Core provider management
    ProviderRegistry() provider.Registry
    
    // Path resolution
    PathResolver() resolver.PathDefinitionResolver
    
    // Upload orchestration (delegates to upload services)
    UploadManager() upload.UploadServiceFacade
    
    // Lifecycle
    Close() error
}
```


### Phase 3: Enhance Upload Aggregate

#### 3.1 Add Resource Value Field
```go
// upload/upload.go - Enhanced Upload struct

type Upload struct {
    // ... existing fields ...
    
    // NEW: CDN path context from business entity
    ResourceValue string `json:"resource_value,omitempty"` 
    ResourceProvider provider.ProviderName `json:"resource_provider,omitempty"` // e.g., 'cdn', 's3', etc.
    
    // Enhanced ETags structure for unified handling
    StorageETags json.RawMessage `json:"storage_etags,omitempty"` 
    // Simple: "abc123" 
    // Multipart: [{"part": 1, "etag": "abc123"}, {"part": 2, "etag": "def456"}]
}
```

#### 3.2 Enhance Upload Aggregate with Rich Behavior
```go
// upload/upload_aggregate.go (enhance existing upload.go)

// Add rich behavior methods to existing Upload struct
func (u *Upload) Start() error {
    if !u.canStart() {
        return ErrInvalidStateTransition
    }
    return u.TransitionStatus(UploadStatusUploading)
}

func (u *Upload) Complete(confirmation *UploadConfirmation) error {
    if !u.canComplete() {
        return ErrInvalidStateTransition
    }
    
    // Enhanced validation logic
    if err := u.validateCompletion(confirmation); err != nil {
        return u.Fail(err.Error())
    }
    
    // Update aggregate state
    u.StorageSize = confirmation.FileSize
    if confirmation.Metadata != nil {
        u.StorageMetadata = confirmation.Metadata
    }
    
    return u.TransitionStatus(UploadStatusCompleted)
}

func (u *Upload) Fail(reason string) error {
    u.Error = &UploadError{
        Code:    "upload_failed",
        Message: reason,
        Time:    time.Now(),
    }
    return u.TransitionStatus(UploadStatusFailed)
}

// Enhanced validation
func (u *Upload) validateCompletion(confirmation *UploadConfirmation) error {
    if !confirmation.Success {
        return fmt.Errorf("upload not successful: %s", confirmation.Error)
    }
    
    if u.Type == UploadTypeMultipart {
        if len(confirmation.ETags) != u.TotalParts {
            return ErrPartCountMismatch.Withf("expected %d parts, got %d", u.TotalParts, len(confirmation.ETags))
        }
    }
    
    return nil
}

func (u *Upload) canStart() bool {
    return u.Status == UploadStatusPending || u.Status == UploadStatusInitializing
}

func (u *Upload) canComplete() bool {
    return u.Status == UploadStatusUploading || u.Status == UploadStatusPending
}
```

## Implementation Timeline

### Phase 1: Database & Model Alignment (Week 1)
- [ ] Update Upload aggregate with resource_value field
- [ ] Enhance ETags storage structure
- [ ] Update database schema to match flow document
- [ ] Add migration scripts for existing data

### Phase 2: Service Separation (Week 2-3)
- [ ] Create SimpleUploadService interface and implementation
- [ ] Create MultipartUploadService interface and implementation
- [ ] Implement UploadServiceFacade
- [ ] Maintain backward compatibility with existing Manager

### Phase 3: Upload Flow Implementation (Week 4-5)
- [ ] Implement deterministic path generation
- [ ] Add upload confirmation workflow
- [ ] Implement cleanup job for expired uploads
- [ ] Add retry mechanism with same path

### Phase 4: Testing & Documentation (Week 6)
- [ ] Write comprehensive unit tests
- [ ] Add integration tests for complete flows
- [ ] Update API documentation
- [ ] Create migration guide

## Migration Strategy

### Backward Compatibility
1. Keep existing `Manager` interface as deprecated facade
2. Implement it using new services internally
3. Add deprecation notices with migration timeline
4. Provide migration helper functions

### Gradual Migration Path
```go
// Deprecated: Use UploadServiceFacade instead
type Manager interface {
    // ... existing interface ...
}

// Implementation delegates to new services
type managerAdapter struct {
    facade UploadServiceFacade
}

func (m *managerAdapter) InitiateUpload(ctx context.Context, opts *UploadOptions) (*Upload, *ResolvedResource, error) {
    // Log deprecation warning
    log.Warn("Manager.InitiateUpload is deprecated, use UploadServiceFacade")
    return m.facade.InitiateUpload(ctx, opts)
}
```

## Testing Strategy

### Unit Tests
- Test each service in isolation
- Mock dependencies
- Focus on business logic
- Test state transitions
- Test aggregate behavior methods

### Integration Tests
- Test service coordination
- Test with real providers (test environment)
- Test error scenarios
- Performance benchmarks
- Concurrent upload scenarios

### Example Test Structure
```go
func TestSimpleUploadService_CompleteFlow(t *testing.T) {
    // Setup
    repo := mocks.NewMockRepository()
    provider := mocks.NewMockProvider()
    service := NewSimpleUploadService(repo, provider)
    
    // Test upload flow
    upload, url, err := service.InitiateUpload(ctx, opts)
    assert.NoError(t, err)
    assert.Equal(t, UploadStatusPending, upload.Status)
    
    // Confirm upload
    err = service.ConfirmUpload(ctx, upload.ID, &UploadConfirmation{
        Success: true,
        FileSize: 1024,
    })
    assert.NoError(t, err)
    
    // Verify state
    upload, err = repo.Get(ctx, upload.ID)
    assert.Equal(t, UploadStatusCompleted, upload.Status)
}
```

## Success Metrics

1. **Code Quality**
   - Reduced cyclomatic complexity (target: <10 per method)
   - Improved test coverage (target: >80%)
   - Clear separation of concerns

2. **Performance**
   - Reduced latency for upload initiation (<100ms)
   - Improved throughput for concurrent uploads
   - Efficient memory usage

3. **Maintainability**
   - Easier to add new upload types
   - Clear domain boundaries
   - Reduced coupling between components

4. **Developer Experience**
   - Intuitive API design
   - Comprehensive documentation
   - Clear migration path

## Risk Mitigation

### Technical Risks
1. **Breaking Changes**
   - Mitigation: Maintain backward compatibility with adapters
   - Provide migration tools and guides

2. **Performance Regression**
   - Mitigation: Benchmark before and after
   - Implement caching strategically

3. **Data Migration**
   - Mitigation: No database schema changes required
   - Services work with existing data structures

### Operational Risks
1. **Service Disruption**
   - Mitigation: Feature flags for gradual rollout
   - Canary deployment strategy

2. **Monitoring Gaps**
   - Mitigation: Add metrics and logging
   - Create dashboards for new services

## Conclusion

This implementation plan addresses the core issues in the resource package while maintaining DDD principles in a flat package structure. The phased approach ensures backward compatibility while gradually improving the architecture. The focus on domain modeling, separation of concerns, and proper abstractions will result in a more maintainable and extensible codebase.