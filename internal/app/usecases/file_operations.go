package usecases

import (
	"fmt"

	"avironactive.com/common/context"
	"avironactive.com/resource/provider"

	"avironactive.com/resource"
	"github.com/anh-nguyen/resource-server/internal/app/dto"
)

// FileOperationsUseCase handles file operation use cases
type FileOperationsUseCase struct {
	manager resource.ResourceManager
}

// NewFileOperationsUseCase creates a new file operations use case
func NewFileOperationsUseCase(manager resource.ResourceManager) *FileOperationsUseCase {
	return &FileOperationsUseCase{
		manager: manager,
	}
}

// ListFiles lists files in a resource path with pagination
func (uc *FileOperationsUseCase) ListFiles(ctx context.Context, providerReq, definition string, maxKeys int32, continuationToken, prefix string) (*dto.FileListResponse, error) {
	// Create list request
	listReq := &provider.ListObjectsOptions{
		MaxKeys:           &maxKeys,
		ContinuationToken: &continuationToken,
		Prefix:            &prefix,
	}

	// List objects using the resource manager
	result, err := uc.manager.ListObjects(ctx, provider.ProviderName(providerReq), definition, listReq)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	// Convert to response DTO
	files := make([]dto.FileInfo, 0, len(result.Objects))
	for _, obj := range result.Objects {
		file := dto.FileInfo{
			Key:          obj.Key,
			Size:         obj.Size,
			ETag:         obj.ETag,
			LastModified: *obj.LastModified,
		}
		files = append(files, file)
	}

	var nextContinuationToken string
	if result.NextContinuationToken != nil {
		nextContinuationToken = *result.NextContinuationToken
	}

	response := &dto.FileListResponse{
		Files:             files,
		ContinuationToken: nextContinuationToken,
		IsTruncated:       result.IsTruncated,
		MaxKeys:           int(maxKeys),
	}

	return response, nil
}

// GenerateUploadURL generates a signed URL for file upload
func (uc *FileOperationsUseCase) GenerateUploadURL(ctx context.Context, provider, definition string, req *dto.UploadRequest) (*dto.SignedURLResponse, error) {
	// Create upload options
	opts := req.To()

	signedURL, err := uc.manager.PathDefinitionResolver().ResolveUploadURL(ctx, resource.PathDefinitionName(definition), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to generate upload URL: %w", err)
	}

	// Convert to response DTO
	response := &dto.SignedURLResponse{
		URL:       signedURL.URL.URL,
		Method:    signedURL.URL.Method,
		Headers:   signedURL.URL.Headers,
		ExpiresAt: signedURL.URL.ExpiresAt,
	}

	return response, nil
}

// GenerateDownloadURL generates a signed URL for file download
func (uc *FileOperationsUseCase) GenerateDownloadURL(ctx context.Context, provider, filePath string, req *dto.DownloadRequest) (*dto.SignedURLResponse, error) {
	// Generate signed URL
	signedURL, err := uc.manager.PathURLResolver().ResolveDownloadURL(ctx, filePath, &resource.DownloadResolveOptions{})
	if err != nil {
		return nil, fmt.Errorf("failed to generate download URL: %w", err)
	}

	// Convert to response DTO
	response := &dto.SignedURLResponse{
		URL:       signedURL.URL.URL,
		Method:    signedURL.URL.Method,
		Headers:   signedURL.URL.Headers,
		ExpiresAt: signedURL.URL.ExpiresAt,
	}

	return response, nil
}

// GenerateMultipartUploadURLs generates URLs for multipart upload workflow
func (uc *FileOperationsUseCase) GenerateMultipartUploadURLs(ctx context.Context, providerReq, definition string, req *dto.MultipartUploadRequest) (*dto.MultipartUploadResponse, error) {
	return nil, fmt.Errorf("GenerateMultipartUploadURLs not implemented yet")
}

// DeleteFile deletes a file from storage
func (uc *FileOperationsUseCase) DeleteFile(ctx context.Context, providerReq, filePath string) error {
	err := uc.manager.DeleteObject(ctx, provider.ProviderName(providerReq), filePath)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// GetFileMetadata retrieves metadata for a specific file
func (uc *FileOperationsUseCase) GetFileMetadata(ctx context.Context, providerReq, filePath string) (*dto.FileMetadata, error) {
	// Get object metadata
	metadata, err := uc.manager.GetObjectMetadata(ctx, provider.ProviderName(providerReq), filePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file metadata: %w", err)
	}

	// Convert checksums
	checksums := make([]dto.ChecksumInfo, 0, len(metadata.Checksums))
	for _, checksum := range metadata.Checksums {
		checksums = append(checksums, dto.ChecksumInfo{
			Algorithm: string(checksum.Algorithm),
			Value:     checksum.Value,
		})
	}

	// Convert storage class and ACL to strings
	var storageClass string
	if metadata.StorageClass != "" {
		storageClass = string(metadata.StorageClass)
	}

	var acl string
	if metadata.ACL != "" {
		acl = string(metadata.ACL)
	}

	// Convert to response DTO
	response := &dto.FileMetadata{
		Key:                metadata.Key,
		Size:               metadata.Size,
		ContentType:        metadata.ContentType,
		ETag:               metadata.ETag,
		Created:            metadata.Created,
		LastModified:       metadata.LastModified,
		StorageClass:       storageClass,
		CacheControl:       metadata.CacheControl,
		ContentEncoding:    metadata.ContentEncoding,
		ContentDisposition: metadata.ContentDisposition,
		ContentLanguage:    metadata.ContentLanguage,
		Metadata:           metadata.Metadata,
		Checksums:          checksums,
		ACL:                acl,
		ExpirationTime:     metadata.ExpirationTime,
	}

	return response, nil
}

// UpdateFileMetadata updates metadata for an existing file
func (uc *FileOperationsUseCase) UpdateFileMetadata(ctx context.Context, providerReq, filePath string, req *dto.MetadataUpdateRequest) (*dto.FileMetadata, error) {
	// Create metadata update options
	updateOpts := &provider.UpdateMetadata{}

	// Set content type if provided
	if req.ContentType != "" {
		updateOpts.ContentType = &req.ContentType
	}

	// Set content encoding if provided
	if req.ContentEncoding != "" {
		updateOpts.ContentEncoding = &req.ContentEncoding
	}

	// Set content language if provided
	if req.ContentLanguage != "" {
		updateOpts.ContentLanguage = &req.ContentLanguage
	}

	// Set content disposition if provided
	if req.ContentDisposition != "" {
		updateOpts.ContentDisposition = &req.ContentDisposition
	}

	// Set cache control if provided
	if req.CacheControl != "" {
		updateOpts.CacheControl = &req.CacheControl
	}

	// Set ACL if provided
	if req.ACL != "" {
		acl := provider.ACLType(req.ACL)
		updateOpts.ACL = &acl
	}

	// Set custom headers if provided
	if req.CustomHeaders != nil {
		updateOpts.CustomHeaders = req.CustomHeaders
	}

	// Update object metadata
	err := uc.manager.UpdateObjectMetadata(ctx, provider.ProviderName(providerReq), filePath, updateOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to update file metadata: %w", err)
	}

	// Return updated metadata
	return uc.GetFileMetadata(ctx, providerReq, filePath)
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
