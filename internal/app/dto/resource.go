package dto

import (
	"time"

	"avironactive.com/resource"
	"avironactive.com/resource/provider"
)

// PathDefinitionResponse represents a resource path definition in API responses
type PathDefinitionResponse struct {
	Name          string                  `json:"name"`
	DisplayName   string                  `json:"displayName"`
	Description   string                  `json:"description"`
	AllowedScopes []string                `json:"allowedScopes"`
	Parameters    []PathParameterResponse `json:"parameters"`
	Providers     []string                `json:"providers"`
}

// PathParameterResponse represents a path parameter in API responses
type PathParameterResponse struct {
	Name         string   `json:"name"`
	Rules        []string `json:"rules,omitempty"`
	Description  string   `json:"description,omitempty"`
	DefaultValue string   `json:"defaultValue,omitempty"`
}

// ProviderResponse represents a storage provider in API responses
type ProviderResponse struct {
	Name         string               `json:"name"`
	Capabilities ProviderCapabilities `json:"capabilities"`
	Constraints  map[string]any       `json:"constraints,omitempty"`
}

// MultipartCapabilities represents multipart upload capabilities
type MultipartCapabilities struct {
	MinPartSize int64 `json:"minPartSize"` // Minimum allowed part size (e.g., 5MB for S3/R2)
	MaxPartSize int64 `json:"maxPartSize"` // Maximum allowed part size (e.g., 5GB for S3/R2)
	MaxParts    int   `json:"maxParts"`    // Maximum number of parts allowed (e.g., 10000 for S3/R2)
}

// ProviderCapabilities represents capabilities of a storage provider
type ProviderCapabilities struct {
	SupportsRead               bool                         `json:"supportsRead"`               // Supports reading objects
	SupportsWrite              bool                         `json:"supportsWrite"`              // Supports writing objects
	SupportsDelete             bool                         `json:"supportsDelete"`             // Supports deleting objects
	SupportsListing            bool                         `json:"supportsListing"`            // Supports listing objects with pagination
	SupportsMetadata           bool                         `json:"supportsMetadata"`           // Supports retrieving and updating object metadata
	SupportsMultipart          bool                         `json:"supportsMultipart"`          // Supports multipart uploads
	SupportsResumableUploads   bool                         `json:"supportsResumableUploads"`   // Supports resumable uploads (if different from multipart)
	SupportsSignedURLs         bool                         `json:"supportsSignedUrls"`         // Supports generating signed URLs for read/write operations
	SupportsChecksumAlgorithms []provider.ChecksumAlgorithm `json:"supportsChecksumAlgorithms"` // Supported checksum algorithms for uploads and downloads
	MaxUploadSize              int64                        `json:"maxUploadSize"`              // Maximum single object size (for non-multipart uploads)
	MaxExpiry                  time.Duration                `json:"maxExpiry"`                  // Maximum allowed expiry for signed URLs
	MinExpiry                  time.Duration                `json:"minExpiry"`                  // Minimum allowed expiry for signed URLs
	Multipart                  *MultipartCapabilities       `json:"multipart,omitempty"`        // Multipart upload capabilities (if supported)
}

// FileListResponse represents a paginated list of files
type FileListResponse struct {
	Files             []FileInfo `json:"files"`
	ContinuationToken string     `json:"continuationToken,omitempty"`
	IsTruncated       bool       `json:"isTruncated"`
	MaxKeys           int        `json:"maxKeys"`
}

// FileInfo represents file information in list responses
type FileInfo struct {
	Key          string         `json:"key"`
	Size         int64          `json:"size"`
	ContentType  string         `json:"contentType,omitempty"`
	ETag         string         `json:"etag,omitempty"`
	LastModified time.Time      `json:"lastModified"`
	Metadata     map[string]any `json:"metadata,omitempty"`
}

// UploadRequest represents a request to generate upload URL
type UploadRequest struct {
	Parameters map[string]string `json:"parameters" validate:"dive,keys,alphanum,endkeys,max=256"`
	Scope      string            `json:"scope" validate:"omitempty,oneof=G A CA"`
	ScopeValue int               `json:"scopeValue,omitempty" validate:"omitempty,min=1"`
	Expiry     string            `json:"expiry,omitempty" validate:"omitempty,duration"`
	Metadata   UploadMetadata    `json:"metadata,omitempty" validate:"omitempty"`
}

func (r UploadRequest) To() *resource.UploadResolveOptions {
	scope := parseScope(r.Scope)
	params := make(map[resource.ParameterName]string)
	for k, v := range r.Parameters {
		params[resource.ParameterName(k)] = v
	}

	return (&resource.UploadResolveOptions{}).WithValues(params).WithScope(scope, r.ScopeValue)
}

// parseScope converts string scope to resource.Scope
func parseScope(scope string) resource.ScopeType {
	switch scope {
	case "G":
		return resource.ScopeGlobal
	case "A":
		return resource.ScopeApp
	case "CA":
		return resource.ScopeClientApp
	default:
		return resource.ScopeGlobal
	}
}

// UploadMetadata represents metadata for upload operations
type UploadMetadata struct {
	ContentType        string            `json:"contentType,omitempty" validate:"omitempty,max=128"`
	ContentEncoding    string            `json:"contentEncoding,omitempty" validate:"omitempty,max=64"`
	ContentLanguage    string            `json:"contentLanguage,omitempty" validate:"omitempty,max=32"`
	ContentDisposition string            `json:"contentDisposition,omitempty" validate:"omitempty,max=256"`
	CacheControl       string            `json:"cacheControl,omitempty" validate:"omitempty,max=128"`
	ACL                string            `json:"acl,omitempty" validate:"omitempty,oneof=private public-read public-read-write authenticated-read"`
	CustomHeaders      map[string]string `json:"customHeaders,omitempty" validate:"omitempty,dive,keys,max=64,endkeys,max=512"`
}

// MultipartUploadRequest represents a request for multipart upload
type MultipartUploadRequest struct {
	Parameters map[string]string `json:"parameters" validate:"dive,keys,alphanum,endkeys,max=256"`
	Scope      string            `json:"scope" validate:"omitempty,oneof=G A CA"`
	ScopeValue int               `json:"scopeValue,omitempty" validate:"omitempty,min=1"`
	FileSize   int64             `json:"fileSize" validate:"required,min=1,max=5368709120"` // 5GB max
	PartCount  int               `json:"partCount" validate:"required,min=1,max=10000"`
	Metadata   *UploadMetadata   `json:"metadata,omitempty"`
}

// DownloadRequest represents a request to generate download URL
type DownloadRequest struct {
	Expiry          string            `json:"expiry,omitempty" validate:"omitempty,duration"`
	ResponseHeaders map[string]string `json:"responseHeaders,omitempty" validate:"omitempty,dive,keys,max=64,endkeys,max=512"`
}

// SignedURLResponse represents a signed URL response
type SignedURLResponse struct {
	URL       string            `json:"url"`
	Method    string            `json:"method"`
	Headers   map[string]string `json:"headers,omitempty"`
	ExpiresAt time.Time         `json:"expiresAt"`
}

// MultipartUploadResponse represents a multipart upload response
type MultipartUploadResponse struct {
	UploadID    string             `json:"uploadId"`
	PartURLs    []MultipartPartURL `json:"partUrls"`
	CompleteURL string             `json:"completeUrl"`
	AbortURL    string             `json:"abortUrl"`
	MinPartSize int64              `json:"minPartSize"`
	MaxPartSize int64              `json:"maxPartSize"`
}

// MultipartPartURL represents a single part URL in multipart upload
type MultipartPartURL struct {
	PartNumber int    `json:"partNumber"`
	URL        string `json:"url"`
	Method     string `json:"method"`
}

// FileMetadata represents file metadata
type FileMetadata struct {
	Key                string            `json:"key"`
	Size               int64             `json:"size"`
	ContentType        string            `json:"contentType,omitempty"`
	ETag               string            `json:"etag,omitempty"`
	Created            *time.Time        `json:"created,omitempty"`
	LastModified       *time.Time        `json:"lastModified,omitempty"`
	StorageClass       string            `json:"storageClass,omitempty"`
	CacheControl       string            `json:"cacheControl,omitempty"`
	ContentEncoding    string            `json:"contentEncoding,omitempty"`
	ContentDisposition string            `json:"contentDisposition,omitempty"`
	ContentLanguage    string            `json:"contentLanguage,omitempty"`
	Metadata           map[string]string `json:"metadata,omitempty"`
	Checksums          []ChecksumInfo    `json:"checksums,omitempty"`
	ACL                string            `json:"acl,omitempty"`
	ExpirationTime     *time.Time        `json:"expirationTime,omitempty"`
}

// ChecksumInfo represents checksum information
type ChecksumInfo struct {
	Algorithm string `json:"algorithm"`
	Value     string `json:"value"`
}

// MetadataUpdateRequest represents a request to update file metadata
type MetadataUpdateRequest struct {
	ContentType        string            `json:"contentType,omitempty" validate:"omitempty,max=128"`
	ContentEncoding    string            `json:"contentEncoding,omitempty" validate:"omitempty,max=64"`
	ContentLanguage    string            `json:"contentLanguage,omitempty" validate:"omitempty,max=32"`
	ContentDisposition string            `json:"contentDisposition,omitempty" validate:"omitempty,max=256"`
	CacheControl       string            `json:"cacheControl,omitempty" validate:"omitempty,max=128"`
	ACL                string            `json:"acl,omitempty" validate:"omitempty,oneof=private public-read public-read-write authenticated-read"`
	CustomHeaders      map[string]string `json:"customHeaders,omitempty" validate:"omitempty,dive,keys,max=64,endkeys,max=512"`
}
