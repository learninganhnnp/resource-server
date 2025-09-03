# Product Requirements Document: Signed URL Generation Feature

## Executive Summary
This document outlines the requirements for implementing a comprehensive signed URL generation system that supports both read and write operations across R2 and GCS storage providers. The system will enable secure, time-limited access to storage resources with support for single uploads and multipart uploads (R2 only).

## 1. Current State Analysis

### Existing Infrastructure
- **Providers**: R2Provider and GCSProvider already implemented with basic signed URL support
- **Current Capabilities**:
  - R2: Currently generates presigned PUT URLs (line 107-118 in r2_provider.go)
  - GCS: Generates signed GET URLs (line 87-98 in gcs_provider.go)
  - Both providers have multipart upload infrastructure (R2 only has implementation)
  - URLOptions struct exists with SignedURL flag and SignedExpiry duration

### Gaps Identified
1. No unified interface for signed URL generation with different HTTP methods
2. Missing support for custom headers in signed URLs
3. No structured response format for multipart upload URLs
4. Lack of comprehensive write URL generation for both providers
5. No abstraction for handling provider-specific signing requirements

## 2. Requirements

### 2.1 Functional Requirements

#### 2.1.1 Read Signed URLs
- **Input Parameters**:
  - `path`: Object storage path
  - `expiry`: Custom expiration duration (optional, defaults to provider config)
  - `signedURLOptions`: Additional options for URL generation
  
- **Output**:
  - `signedURL`: Time-limited URL for reading the object
  - `expiresAt`: Timestamp when the URL expires

#### 2.1.2 Write Signed URLs - Single Upload

**Supported Providers**: R2, GCS

- **Input Parameters**:
  - `path`: Target object storage path
  - `expiry`: Custom expiration duration
  - `headers`: Custom HTTP headers (Content-Type, Cache-Control, etc.)
  
- **Output**:
  - `signedURL`: Pre-signed URL for direct upload
  - `headers`: Required headers to include in the upload request
  - `method`: HTTP method to use (PUT/POST)
  - `expiresAt`: Timestamp when the URL expires

#### 2.1.3 Write Signed URLs - Multipart Upload

**Supported Providers**: R2 only

**CRITICAL: Two-Phase Approach Required**

Due to signature verification requirements, multipart uploads MUST be handled in two phases to avoid signature invalidation issues.

##### Phase 1: Initiation URLs (No Upload ID)

- **Input Parameters**:
  - `path`: Target object storage path
  - `expiry`: Custom expiration duration
  - `headers`: Custom HTTP headers
  - `fileSize`: Total size of the file to upload (for calculating parts)
  
- **Output Structure**:
```go
type MultipartInitResponse struct {
    // URL to initiate multipart upload (no upload ID needed)
    InitiateURL   string            
    
    // Configuration for the upload
    ChunkSize     int64             // Recommended size for each part (e.g., 10MB)
    TotalParts    int               // Total number of parts expected
    MinPartSize   int64             // Minimum allowed part size (5MB for S3/R2)
    MaxPartSize   int64             // Maximum allowed part size (5GB for S3/R2)
    
    // Headers to include in initiate request
    Headers       map[string]string 
    
    // Expiration
    ExpiresAt     time.Time         
    
    // Instructions for next step
    NextStep      string            // "Call initiate URL to get upload ID, then request part URLs"
}
```

##### Phase 2: Part/Complete/Abort URLs (With Real Upload ID)

- **Input Parameters**:
  - `path`: Target object storage path
  - `uploadID`: Real upload ID received from R2/S3
  - `partNumbers`: List of part numbers needing URLs (or range)
  - `expiry`: Custom expiration duration
  
- **Output Structure**:
```go
type MultipartURLsResponse struct {
    // Upload identifier
    UploadID      string            // Echo back the provided upload ID
    
    // URLs for multipart operations (all with real upload ID embedded)
    PartURLs      map[int]PartURL   // Map of part number to URL details
    CompleteURL   string            // URL to complete the upload
    AbortURL      string            // URL to abort the upload
    
    // Required headers for each operation
    PartHeaders   map[string]string // Headers for part uploads
    
    // Expiration
    ExpiresAt     time.Time         
}

type PartURL struct {
    PartNumber int
    URL        string
    MinSize    int64  // Minimum size for this part (5MB except last part)
    MaxSize    int64  // Maximum size for this part
}
```

### 2.2 Technical Requirements

#### 2.2.1 Architecture Principles
- **SOLID Compliance**:
  - Single Responsibility: Separate URL generation from storage operations
  - Open/Closed: Extensible for new providers without modifying existing code
  - Liskov Substitution: All providers implement consistent interfaces
  - Interface Segregation: Separate interfaces for read/write operations
  - Dependency Inversion: Depend on abstractions, not concrete implementations

#### 2.2.2 Provider Compatibility
- **R2 Provider**:
  - Use AWS SDK v2 with v4.Signer for signing
  - Support S3-compatible multipart upload API
  - Handle R2-specific endpoint formatting
  - CRITICAL: Never use placeholder values in signed URLs
  
- **GCS Provider**:
  - Use Google Cloud Storage client library
  - Support V4 signing scheme
  - Handle service account and ADC authentication
  - Note: GCS uses resumable uploads instead of multipart

#### 2.2.3 Signature Integrity
- **No Placeholders**: Signed URLs must NEVER contain placeholder values
- **Exact Matching**: Every character in the signed URL must match what was signed
- **Fresh Generation**: Generate new signatures for each request with actual values
- **No String Replacement**: Client must never modify signed URLs

### 2.3 Non-Functional Requirements

#### 2.3.1 Security
- All signed URLs must have expiration times
- Support for custom headers to enforce content restrictions
- No exposure of credentials in generated URLs
- Secure handling of signing keys and service accounts
- Prevent signature tampering through proper validation

#### 2.3.2 Performance
- URL generation should be fast (<100ms)
- Support for concurrent URL generation
- Efficient calculation of multipart chunks
- Consider caching upload sessions server-side

#### 2.3.3 Reliability
- Graceful error handling for invalid inputs
- Clear error messages for debugging
- Fallback to default configurations when optional parameters are missing
- Proper handling of signature verification failures

## 3. Implementation Design

### 3.1 Interface Design

```go
// SignedURLGenerator defines the interface for generating signed URLs
type SignedURLGenerator interface {
    // Generate read URL
    GenerateReadURL(ctx context.Context, path string, opts *ReadURLOptions) (*SignedURLResponse, error)
    
    // Generate write URL for single upload
    GenerateWriteURL(ctx context.Context, path string, opts *WriteURLOptions) (*SignedURLResponse, error)
    
    // Multipart upload support (R2 only)
    MultipartUploader
}

// MultipartUploader handles multipart upload URL generation
type MultipartUploader interface {
    // Phase 1: Generate initiation URL (no upload ID required)
    GenerateMultipartInitURL(ctx context.Context, path string, opts *MultipartInitOptions) (*MultipartInitResponse, error)
    
    // Phase 2: Generate part URLs with real upload ID
    GenerateMultipartURLs(ctx context.Context, path string, uploadID string, opts *MultipartURLOptions) (*MultipartURLsResponse, error)
}

// ReadURLOptions configures read URL generation
type ReadURLOptions struct {
    Expiry          time.Duration
    ResponseHeaders map[string]string // Headers to override in response
}

// WriteURLOptions configures write URL generation
type WriteURLOptions struct {
    Expiry        time.Duration
    Headers       *ObjectHeaders
    StorageClass  string
    ACL           string
}

// MultipartInitOptions for initiating multipart upload
type MultipartInitOptions struct {
    Expiry       time.Duration
    Headers      *ObjectHeaders
    StorageClass string
    ACL          string
    FileSize     int64  // Total file size for calculating parts
    PartSize     int64  // Optional: Override default part size
}

// MultipartURLOptions for generating part URLs
type MultipartURLOptions struct {
    Expiry      time.Duration
    PartNumbers []int  // Which parts need URLs
    // Or use range
    StartPart   int    // Generate URLs from this part
    EndPart     int    // Generate URLs up to this part
}

// SignedURLResponse represents a generated signed URL
type SignedURLResponse struct {
    URL         string
    Method      string            // HTTP method to use
    Headers     map[string]string // Required headers for the request
    ExpiresAt   time.Time
}
```

### 3.2 Flow Diagrams

#### 3.2.1 Single Upload Flow (R2/GCS)
```
Client                  Server                  Storage Provider
  |                       |                           |
  |--Request Write URL--->|                           |
  |                       |                           |
  |                       |--Generate Signed URL----->|
  |                       |<--Return Signed URL-------|
  |                       |                           |
  |<--Signed URL+Headers--|                           |
  |                       |                           |
  |--Upload with URL+Headers------------------------->|
  |<--Upload Success----------------------------------|
```

#### 3.2.2 Multipart Upload Flow (R2) - Correct Two-Phase Approach
```
Phase 1: Initiation
Client                  Server                  R2 Storage
  |                       |                           |
  |--Request Init URL---->|                           |
  |  (path, fileSize)     |                           |
  |                       |                           |
  |                       |--Generate Init URL------->|
  |                       |  (no uploadID)            |
  |                       |                           |
  |<--Init URL+Config------|                           |
  |                       |                           |
  |--POST Init URL-------------------------------------->|
  |<--Real UploadID: "abc123"---------------------------|
  |                       |                           |

Phase 2: Part URLs with Real Upload ID
  |                       |                           |
  |--Request Part URLs---->|                           |
  |  (uploadID="abc123")   |                           |
  |                       |                           |
  |                       |--Generate Part URLs------>|
  |                       |  (with real uploadID)     |
  |                       |--Generate Complete URL--->|
  |                       |  (with real uploadID)     |
  |                       |--Generate Abort URL------>|
  |                       |  (with real uploadID)     |
  |                       |                           |
  |<--All URLs with--------|                           |
  |   real uploadID        |                           |
  |                       |                           |
  |--Upload Parts--------------------------------------->|
  |  (using signed URLs)                                |
  |<--Part ETags----------------------------------------|
  |                       |                           |
  |--Complete Upload------------------------------------>|
  |  (using complete URL)                              |
  |<--Success--------------------------------------------|
```

### 3.3 Implementation Examples

#### 3.3.1 R2 Multipart Implementation

```go
// Phase 1: Generate initiation URL without upload ID
func (p *R2Provider) GenerateMultipartInitURL(ctx context.Context, path string, opts *MultipartInitOptions) (*MultipartInitResponse, error) {
    bucketName, actualPath := p.extractBucketAndPath(path)
    
    input := &s3.CreateMultipartUploadInput{
        Bucket: aws.String(bucketName),
        Key:    aws.String(actualPath),
    }
    
    // Apply headers if provided
    if opts.Headers != nil {
        if opts.Headers.ContentType != "" {
            input.ContentType = aws.String(opts.Headers.ContentType)
        }
        if opts.Headers.CacheControl != "" {
            input.CacheControl = aws.String(opts.Headers.CacheControl)
        }
        // ... other headers
    }
    
    // Generate presigned URL for initiation
    presignedReq, err := p.presigner.PresignCreateMultipartUpload(ctx, input, func(po *s3.PresignOptions) {
        po.Expires = opts.Expiry
    })
    
    if err != nil {
        return nil, err
    }
    
    // Calculate part configuration
    chunkSize := calculateOptimalChunkSize(opts.FileSize, opts.PartSize)
    totalParts := calculateTotalParts(opts.FileSize, chunkSize)
    
    return &MultipartInitResponse{
        InitiateURL: presignedReq.URL,
        ChunkSize:   chunkSize,
        TotalParts:  totalParts,
        MinPartSize: 5 * 1024 * 1024,  // 5MB minimum for R2/S3
        MaxPartSize: 5 * 1024 * 1024 * 1024, // 5GB maximum
        Headers:     presignedReq.SignedHeader,
        ExpiresAt:   time.Now().Add(opts.Expiry),
        NextStep:    "POST to InitiateURL to get UploadID, then request part URLs",
    }, nil
}

// Phase 2: Generate URLs with real upload ID
func (p *R2Provider) GenerateMultipartURLs(ctx context.Context, path string, uploadID string, opts *MultipartURLOptions) (*MultipartURLsResponse, error) {
    if uploadID == "" {
        return nil, anerror.ErrInvalidArgument.With("uploadID is required")
    }
    
    bucketName, actualPath := p.extractBucketAndPath(path)
    
    // Determine which parts to generate URLs for
    partNumbers := opts.PartNumbers
    if len(partNumbers) == 0 && opts.EndPart > 0 {
        // Generate range
        for i := opts.StartPart; i <= opts.EndPart; i++ {
            partNumbers = append(partNumbers, i)
        }
    }
    
    // Generate part URLs
    partURLs := make(map[int]PartURL)
    for _, partNum := range partNumbers {
        input := &s3.UploadPartInput{
            Bucket:     aws.String(bucketName),
            Key:        aws.String(actualPath),
            UploadId:   aws.String(uploadID), // Real upload ID
            PartNumber: aws.Int32(int32(partNum)),
        }
        
        presignedReq, err := p.presigner.PresignUploadPart(ctx, input, func(po *s3.PresignOptions) {
            po.Expires = opts.Expiry
        })
        
        if err != nil {
            return nil, err
        }
        
        partURLs[partNum] = PartURL{
            PartNumber: partNum,
            URL:        presignedReq.URL,
        }
    }
    
    // Generate complete URL
    completeInput := &s3.CompleteMultipartUploadInput{
        Bucket:   aws.String(bucketName),
        Key:      aws.String(actualPath),
        UploadId: aws.String(uploadID),
    }
    
    completeReq, err := p.presigner.PresignCompleteMultipartUpload(ctx, completeInput, func(po *s3.PresignOptions) {
        po.Expires = opts.Expiry
    })
    
    if err != nil {
        return nil, err
    }
    
    // Generate abort URL
    abortInput := &s3.AbortMultipartUploadInput{
        Bucket:   aws.String(bucketName),
        Key:      aws.String(actualPath),
        UploadId: aws.String(uploadID),
    }
    
    abortReq, err := p.presigner.PresignAbortMultipartUpload(ctx, abortInput, func(po *s3.PresignOptions) {
        po.Expires = opts.Expiry
    })
    
    if err != nil {
        return nil, err
    }
    
    return &MultipartURLsResponse{
        UploadID:    uploadID,
        PartURLs:    partURLs,
        CompleteURL: completeReq.URL,
        AbortURL:    abortReq.URL,
        ExpiresAt:   time.Now().Add(opts.Expiry),
    }, nil
}

// Helper function to calculate optimal chunk size
func calculateOptimalChunkSize(fileSize, requestedSize int64) int64 {
    const (
        minPartSize = 5 * 1024 * 1024        // 5MB
        maxPartSize = 5 * 1024 * 1024 * 1024 // 5GB
        maxParts    = 10000                   // S3/R2 limit
    )
    
    chunkSize := requestedSize
    if chunkSize == 0 {
        // Auto-calculate based on file size
        chunkSize = fileSize / 100 // Aim for ~100 parts
        
        // Enforce minimum
        if chunkSize < minPartSize {
            chunkSize = minPartSize
        }
        
        // Enforce maximum
        if chunkSize > maxPartSize {
            chunkSize = maxPartSize
        }
        
        // Check if we exceed max parts
        if fileSize/chunkSize > maxParts {
            chunkSize = fileSize / maxParts
            // Round up to ensure we don't exceed max parts
            chunkSize = ((chunkSize / minPartSize) + 1) * minPartSize
        }
    }
    
    return chunkSize
}
```

## 4. Implementation Phases

### Phase 1: Foundation (Priority: High)
1. Define new interfaces and data structures
2. Refactor existing URLOptions to support new requirements
3. Create base implementation for SignedURLGenerator
4. Design two-phase multipart upload flow

### Phase 2: Read URL Implementation (Priority: High)
1. Implement GenerateReadURL for R2Provider
2. Implement GenerateReadURL for GCSProvider
3. Add support for response header overrides
4. Create comprehensive unit tests

### Phase 3: Single Upload Implementation (Priority: High)
1. Implement GenerateWriteURL for R2Provider
2. Implement GenerateWriteURL for GCSProvider
3. Add custom header support
4. Create integration tests

### Phase 4: Multipart Upload Implementation (Priority: Medium)
1. Implement GenerateMultipartInitURL for R2Provider
2. Implement GenerateMultipartURLs for R2Provider
3. Add intelligent chunk size calculation
4. Create helper methods for part management
5. Implement comprehensive error handling
6. Add upload session caching (optional)

### Phase 5: Enhancement & Optimization (Priority: Low)
1. Add request signing for custom headers
2. Implement URL caching for frequently accessed objects
3. Add metrics and monitoring
4. Create performance benchmarks
5. Consider GCS resumable upload support

## 5. Testing Strategy

### 5.1 Unit Tests
- Test URL generation with various input combinations
- Validate expiration time calculations
- Test header merging and validation
- Verify error handling for invalid inputs
- Test signature generation without placeholders

### 5.2 Integration Tests
- Test actual upload/download with generated URLs
- Verify multipart upload flow end-to-end
- Test with different file sizes and types
- Validate cross-provider compatibility
- Verify signature validation with real R2/GCS endpoints

### 5.3 Performance Tests
- Benchmark URL generation speed
- Test concurrent URL generation
- Measure memory usage for large multipart calculations

### 5.4 Security Tests
- Verify no credential leakage in URLs
- Test signature expiration enforcement
- Validate that modified URLs are rejected
- Ensure placeholder replacement doesn't work

## 6. Success Metrics

1. **Functional Completeness**
   - 100% support for read URLs on both providers
   - 100% support for single upload URLs on both providers
   - Full multipart upload support for R2
   - Zero signature validation failures

2. **Performance**
   - URL generation < 100ms for 95% of requests
   - Support for files up to 5TB (R2 multipart)
   - Efficient handling of 10,000 parts

3. **Reliability**
   - Zero critical bugs in production
   - 99.9% success rate for URL generation
   - Clear error messages for all failure scenarios
   - Proper handling of upload ID management

## 7. Dependencies

### External Dependencies
- AWS SDK v2 for R2 operations
- Google Cloud Storage client library
- Standard Go libraries for cryptographic operations

### Internal Dependencies
- Existing provider infrastructure
- Error handling framework (anerror)
- Configuration management system

## 8. Risks and Mitigations

| Risk | Impact | Mitigation |
|------|--------|------------|
| Provider API changes | High | Version lock dependencies, monitor provider changelogs |
| Signature calculation errors | High | Extensive testing, use official SDKs, never use placeholders |
| Performance degradation | Medium | Implement caching, optimize calculations |
| Security vulnerabilities | High | Regular security audits, follow best practices |
| Upload ID management complexity | Medium | Clear documentation, consider server-side session caching |
| Client implementation errors | Medium | Provide client SDK/examples, clear API documentation |

## 9. Client Implementation Guide

### 9.1 Correct Multipart Upload Flow

```javascript
// Client-side JavaScript example
async function uploadLargeFile(file, path) {
    // Phase 1: Get initiation URL
    const initResponse = await fetch('/api/multipart/init', {
        method: 'POST',
        body: JSON.stringify({
            path: path,
            fileSize: file.size
        })
    });
    const { initiateURL, chunkSize, totalParts } = await initResponse.json();
    
    // Phase 2: Initiate upload with R2
    const initResult = await fetch(initiateURL, {
        method: 'POST',
        headers: initResponse.headers
    });
    const uploadId = await extractUploadId(initResult); // Parse from response
    
    // Phase 3: Get part URLs with real upload ID
    const partResponse = await fetch('/api/multipart/parts', {
        method: 'POST',
        body: JSON.stringify({
            path: path,
            uploadId: uploadId,  // Real upload ID from R2
            partNumbers: Array.from({length: totalParts}, (_, i) => i + 1)
        })
    });
    const { partURLs, completeURL } = await partResponse.json();
    
    // Phase 4: Upload parts
    const etags = [];
    for (let i = 0; i < totalParts; i++) {
        const start = i * chunkSize;
        const end = Math.min(start + chunkSize, file.size);
        const chunk = file.slice(start, end);
        
        const uploadResult = await fetch(partURLs[i + 1].url, {
            method: 'PUT',
            body: chunk
        });
        
        etags.push({
            partNumber: i + 1,
            etag: uploadResult.headers.get('ETag')
        });
    }
    
    // Phase 5: Complete upload
    await fetch(completeURL, {
        method: 'POST',
        body: buildCompleteXML(etags)
    });
}
```

### 9.2 Common Mistakes to Avoid

1. **Never modify signed URLs** - Use them exactly as provided
2. **Don't cache signed URLs** - They expire and become invalid
3. **Don't reuse upload IDs** - Each upload needs its own session
4. **Don't mix parts from different uploads** - Keep sessions separate

## 10. Future Enhancements

1. Support for additional providers (Azure Blob, MinIO)
2. Batch URL generation for multiple objects
3. Server-side upload session management
4. Advanced ACL and policy support
5. WebSocket support for real-time uploads
6. Resumable upload support for GCS
7. Automatic retry with fresh URLs on expiration

## 11. Appendix

### A. Code Examples

#### Example 1: Generate Read URL
```go
readOpts := &ReadURLOptions{
    Expiry: 1 * time.Hour,
    ResponseHeaders: map[string]string{
        "Content-Disposition": "attachment; filename=document.pdf",
    },
}

response, err := provider.GenerateReadURL(ctx, "bucket/path/to/file.pdf", readOpts)
if err != nil {
    return err
}

fmt.Printf("URL: %s, Expires: %s\n", response.URL, response.ExpiresAt)
```

#### Example 2: Correct Multipart Upload Usage
```go
// Step 1: Initialize
initOpts := &MultipartInitOptions{
    Expiry:   2 * time.Hour,
    FileSize: 1024 * 1024 * 1024, // 1GB
    Headers: &ObjectHeaders{
        ContentType: "video/mp4",
    },
}

initResp, err := r2Provider.GenerateMultipartInitURL(ctx, "bucket/video.mp4", initOpts)
if err != nil {
    return err
}

// Client calls initiate URL and gets upload ID...
uploadID := "abc123-received-from-r2"

// Step 2: Get part URLs with real upload ID
urlOpts := &MultipartURLOptions{
    Expiry:    2 * time.Hour,
    StartPart: 1,
    EndPart:   initResp.TotalParts,
}

urlResp, err := r2Provider.GenerateMultipartURLs(ctx, "bucket/video.mp4", uploadID, urlOpts)
if err != nil {
    return err
}

// Now client can upload parts using urlResp.PartURLs
```

### B. References

1. [AWS S3 Presigned URLs Documentation](https://docs.aws.amazon.com/AmazonS3/latest/userguide/PresignedUrlUploadObject.html)
2. [AWS Signature Version 4](https://docs.aws.amazon.com/general/latest/gr/signature-version-4.html)
3. [Google Cloud Storage Signed URLs](https://cloud.google.com/storage/docs/access-control/signed-urls)
4. [Cloudflare R2 Documentation](https://developers.cloudflare.com/r2/)
5. [S3 Multipart Upload API](https://docs.aws.amazon.com/AmazonS3/latest/userguide/mpuoverview.html)

---

*Document Version: 2.0*  
*Last Updated: 2025-08-13*  
*Status: Final*  
*Key Update: Corrected multipart upload flow to use two-phase approach without placeholders*