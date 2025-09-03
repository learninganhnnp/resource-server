package dto

import (
	"time"

	"avironactive.com/resource/metadata"
	"avironactive.com/resource/provider"
	"avironactive.com/resource/resolver"
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
}

// NewProviderResponse creates ProviderResponse from provider.Provider
func NewProviderResponse(prov provider.Provider) *ProviderResponse {
	caps := prov.Capabilities()
	return &ProviderResponse{
		Name:         string(prov.Name()),
		Capabilities: *NewProviderCapabilities(caps),
	}
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
	SupportsChecksumAlgorithms []metadata.ChecksumAlgorithm `json:"supportsChecksumAlgorithms"` // Supported checksum algorithms for uploads and downloads
	MaxUploadSize              int64                        `json:"maxUploadSize"`              // Maximum single object size (for non-multipart uploads)
	MaxExpiry                  time.Duration                `json:"maxExpiry"`                  // Maximum allowed expiry for signed URLs
	MinExpiry                  time.Duration                `json:"minExpiry"`                  // Minimum allowed expiry for signed URLs
	Multipart                  *MultipartCapabilities       `json:"multipart,omitempty"`        // Multipart upload capabilities (if supported)
}

func NewProviderCapabilities(caps *provider.Capabilities) *ProviderCapabilities {
	var multipart *MultipartCapabilities
	if caps.Multipart != nil {
		multipart = &MultipartCapabilities{
			MinPartSize: caps.Multipart.MinPartSize,
			MaxPartSize: caps.Multipart.MaxPartSize,
			MaxParts:    caps.Multipart.MaxParts,
		}
	}

	return &ProviderCapabilities{
		SupportsRead:               caps.SupportsRead,
		SupportsWrite:              caps.SupportsWrite,
		SupportsDelete:             caps.SupportsDelete,
		SupportsListing:            caps.SupportsListing,
		SupportsMetadata:           caps.SupportsMetadata,
		SupportsMultipart:          caps.SupportsMultipart,
		SupportsResumableUploads:   caps.SupportsResumableUploads,
		SupportsSignedURLs:         caps.SupportsSignedURLs,
		SupportsChecksumAlgorithms: caps.SupportsChecksumAlgorithms,
		MaxUploadSize:              caps.MaxUploadSize,
		MaxExpiry:                  caps.MaxExpiry,
		MinExpiry:                  caps.MinExpiry,
		Multipart:                  multipart,
	}
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
	ScopeValue int16             `json:"scopeValue,omitempty" validate:"omitempty,min=1"`
	Expiry     string            `json:"expiry,omitempty" validate:"omitempty,duration"`
	Metadata   *UploadMetadata   `json:"metadata,omitempty" validate:"omitempty"`
}

func (r UploadRequest) To() *resolver.DefinitionUploadOptions {
	scope := parseScope(r.Scope)
	params := make(map[resolver.ParameterName]string)
	for k, v := range r.Parameters {
		params[resolver.ParameterName(k)] = v
	}

	opts := &resolver.DefinitionUploadOptions{}
	opts = opts.WithValues(params)
	if r.Scope != "" {
		opts = opts.WithScope(scope, r.ScopeValue)
	}
	return opts
}

// parseScope converts string scope to resolver.Scope
func parseScope(scope string) resolver.ScopeType {
	switch scope {
	case "G":
		return resolver.ScopeGlobal
	case "A":
		return resolver.ScopeApp
	case "CA":
		return resolver.ScopeClientApp
	default:
		return resolver.ScopeGlobal
	}
}

// UploadMetadata represents metadata for upload operations
type UploadMetadata struct {
	ContentType        string            `json:"contentType,omitempty" validate:"omitempty,max=128"`
	ContentEncoding    string            `json:"contentEncoding,omitempty" validate:"omitempty,max=64"`
	ContentLanguage    string            `json:"contentLanguage,omitempty" validate:"omitempty,max=32"`
	ContentDisposition string            `json:"contentDisposition,omitempty" validate:"omitempty,max=256"`
	CacheControl       string            `json:"cacheControl,omitempty" validate:"omitempty,max=128"`
	StorageClass       string            `json:"storageClass,omitempty" validate:"omitempty,max=32,oneof=STANDARD NEARLINE COLDLINE ARCHIVE INTELLIGENT_TIERING"` // Example for GCS
	ACL                string            `json:"acl,omitempty" validate:"omitempty,oneof=private public-read public-read-write authenticated-read"`
	CustomHeaders      map[string]string `json:"customHeaders,omitempty" validate:"omitempty,dive,keys,max=64,endkeys,max=512"`
}

func (m *UploadMetadata) ToRequestHeaders() *metadata.StorageMetadata {
	if m == nil {
		return nil
	}
	return &metadata.StorageMetadata{
		ContentType:        m.ContentType,
		ContentEncoding:    m.ContentEncoding,
		ContentLanguage:    m.ContentLanguage,
		ContentDisposition: m.ContentDisposition,
		CacheControl:       m.CacheControl,
		ACL:                metadata.ACLType(m.ACL),
		StorageClass:       metadata.StorageClassType(m.StorageClass),
		Metadata:           m.CustomHeaders,
	}
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

func (r *DownloadRequest) To() *resolver.DownloadOptions {
	opts := &resolver.DownloadOptions{}
	if r.Expiry != "" {
		// Parse expiry duration and set on options
		// This would need to be implemented based on your duration parsing logic
	}
	if len(r.ResponseHeaders) > 0 {
		// Set response headers on options
		// This would need to be implemented based on your options structure
	}
	return opts
}

// SignedURLResponse represents a signed URL response
type SignedURLResponse struct {
	URL                string            `json:"url"`
	Method             string            `json:"method"`
	Headers            map[string]string `json:"headers,omitempty"`
	ExpiresAt          time.Time         `json:"expiresAt"`
	ResolvedPath       string            `json:"resolvedPath,omitempty"`
	ResolvedParameters map[string]string `json:"resolvedParameters,omitempty"`
}

// MultipartUploadResponse represents a multipart upload response
type MultipartUploadResponse struct {
	UploadID    string             `json:"uploadId"`
	PartURLs    []MultipartPartURL `json:"partUrls"`
	CompleteURL SignedURLResponse  `json:"completeUrl"`
	AbortURL    SignedURLResponse  `json:"abortUrl"`
	MinPartSize int64              `json:"minPartSize"`
	MaxPartSize int64              `json:"maxPartSize"`
}

// MultipartPartURL represents a single part URL in multipart upload
type MultipartPartURL struct {
	SignedURLResponse

	PartNumber int `json:"partNumber"`
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

// ToProviderChecksum converts ChecksumInfo to metadata.Checksum
func (c *ChecksumInfo) ToProviderChecksum() *metadata.Checksum {
	return &metadata.Checksum{
		Algorithm: metadata.ChecksumAlgorithm(c.Algorithm),
		Value:     c.Value,
	}
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

func (r *MetadataUpdateRequest) ToRequestHeaders() *metadata.StorageMetadata {
	return &metadata.StorageMetadata{
		ContentType:        r.ContentType,
		ContentEncoding:    r.ContentEncoding,
		ContentLanguage:    r.ContentLanguage,
		ContentDisposition: r.ContentDisposition,
		CacheControl:       r.CacheControl,
		ACL:                metadata.ACLType(r.ACL),
		Metadata:           r.CustomHeaders,
	}
}

func (r *MetadataUpdateRequest) ToUpdateMetadata() *provider.UpdateMetadata {
	updateOpts := &provider.UpdateMetadata{}

	if r.ContentType != "" {
		updateOpts.ContentType = &r.ContentType
	}
	if r.ContentEncoding != "" {
		updateOpts.ContentEncoding = &r.ContentEncoding
	}
	if r.ContentLanguage != "" {
		updateOpts.ContentLanguage = &r.ContentLanguage
	}
	if r.ContentDisposition != "" {
		updateOpts.ContentDisposition = &r.ContentDisposition
	}
	if r.CacheControl != "" {
		updateOpts.CacheControl = &r.CacheControl
	}
	if r.ACL != "" {
		acl := metadata.ACLType(r.ACL)
		updateOpts.ACL = &acl
	}
	if len(r.CustomHeaders) > 0 {
		updateOpts.CustomHeaders = r.CustomHeaders
	}

	return updateOpts
}

// MultipartInitRequest represents a request to initialize multipart upload
type MultipartInitRequest struct {
	DefinitionName string            `json:"definitionName" validate:"required,alphanum,max=128"`
	Provider       string            `json:"provider" validate:"required,oneof=cdn gcs r2"`
	Scope          string            `json:"scope" validate:"omitempty,oneof=G A CA"`
	ScopeValue     int16             `json:"scopeValue,omitempty" validate:"omitempty,min=1"`
	ParamResolver  map[string]string `json:"paramResolver" validate:"dive,keys,alphanum,endkeys,max=256"`
	Metadata       *UploadMetadata   `json:"metadata,omitempty"`
}

func (req MultipartInitRequest) To() *resolver.DefinitionDownloadOptions {
	opts := &resolver.DefinitionDownloadOptions{}
	provider := provider.ProviderName(req.Provider)
	opts.Provider = &provider

	if req.Scope != "" {
		scope := resolver.ScopeType(req.Scope)
		opts.Scope = &scope
		opts.ScopeValue = req.ScopeValue
	}

	// Add parameters to options
	if req.ParamResolver != nil {
		params := make(map[resolver.ParameterName]string)
		for k, v := range req.ParamResolver {
			params[resolver.ParameterName(k)] = v
		}
		opts = opts.WithValues(params)
	}

	return opts
}

// MultipartInitResponse represents the response from multipart init
type MultipartInitResponse struct {
	UploadID    string `json:"uploadId"`
	Path        string `json:"path"`
	Provider    string `json:"provider"`
	MaxPartSize int64  `json:"maxPartSize,omitempty"`
	MinPartSize int64  `json:"minPartSize,omitempty"`
	MaxParts    int    `json:"maxParts,omitempty"`
}

// MultipartURLsRequest represents a request to get multipart upload URLs
type MultipartURLsRequest struct {
	Path       string         `json:"path" validate:"required,alphanum,max=128"`
	UploadID   string         `json:"uploadId" validate:"required,max=256"`
	Provider   string         `json:"provider" validate:"required,oneof=cdn gcs r2"`
	URLOptions []*PartRequest `json:"urlOptions" validate:"required,dive"`
}

func (req MultipartURLsRequest) To() *resolver.MultipartOptions {
	var parts = make([]provider.Part, 0, len(req.URLOptions))
	for _, partReq := range req.URLOptions {
		parts = append(parts, partReq.To())
	}

	urlOpts := &resolver.MultipartOptions{
		URLOptions: &provider.MultipartURLsOption{
			Parts: parts,
		},
	}
	prov := provider.ProviderName(req.Provider)
	urlOpts.Provider = &prov

	return urlOpts
}

type PartRequest struct {
	PartNumber int          `json:"partNumber" validate:"required,min=1"`
	Checksum   ChecksumInfo `json:"checksum" validate:"required"`
}

func (r *PartRequest) To() provider.Part {
	return provider.Part{
		Number:   r.PartNumber,
		Checksum: r.Checksum.ToProviderChecksum(),
	}
}

// MultipartURLsResponse represents the response with multipart URLs
type MultipartURLsResponse struct {
	PartURLs    []MultipartPartURL `json:"partUrls"`
	CompleteURL SignedURLResponse  `json:"completeUrl"`
	AbortURL    SignedURLResponse  `json:"abortUrl"`
}

// ListFilesRequest represents a request to list files with pagination
type ListFilesRequest struct {
	Provider          string `json:"provider" validate:"required,oneof=cdn gcs r2"`
	Definition        string `json:"definition" validate:"required,alphanum,max=128"`
	MaxKeys           int32  `json:"maxKeys,omitempty" validate:"omitempty,min=1,max=1000"`
	ContinuationToken string `json:"continuationToken,omitempty"`
	Prefix            string `json:"prefix,omitempty" validate:"omitempty,max=256"`
}

// GenerateUploadURLRequest represents a request to generate an upload URL
type GenerateUploadURLRequest struct {
	Provider   string         `json:"provider" validate:"required,oneof=cdn gcs r2"`
	Definition string         `json:"definition" validate:"required,alphanum,max=128"`
	Upload     *UploadRequest `json:"upload" validate:"required"`
}

// GenerateDownloadURLRequest represents a request to generate a download URL
type GenerateDownloadURLRequest struct {
	Provider string           `json:"provider" validate:"required,oneof=cdn gcs r2"`
	FilePath string           `json:"filePath" validate:"required,max=512"`
	Download *DownloadRequest `json:"download,omitempty"`
}

// DeleteFileRequest represents a request to delete a file
type DeleteFileRequest struct {
	Provider string `json:"provider" validate:"required,oneof=cdn gcs r2"`
	FilePath string `json:"filePath" validate:"required,max=512"`
}

// GetFileMetadataRequest represents a request to get file metadata
type GetFileMetadataRequest struct {
	Provider string `json:"provider" validate:"required,oneof=cdn gcs r2"`
	FilePath string `json:"filePath" validate:"required,max=512"`
}

// UpdateFileMetadataRequest represents a request to update file metadata
type UpdateFileMetadataRequest struct {
	Provider string                 `json:"provider" validate:"required,oneof=cdn gcs r2"`
	FilePath string                 `json:"filePath" validate:"required,max=512"`
	Metadata *MetadataUpdateRequest `json:"metadata" validate:"required"`
}
