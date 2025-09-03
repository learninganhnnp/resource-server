package usecases

import (
	"fmt"

	"avironactive.com/common/context"
	"avironactive.com/resource/provider"
	"avironactive.com/resource/resolver"

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
func (uc *FileOperationsUseCase) ListFiles(ctx context.Context, req *dto.ListFilesRequest) (*dto.FileListResponse, error) {
	listReq := &provider.ListObjectsOptions{
		MaxKeys:           &req.MaxKeys,
		ContinuationToken: &req.ContinuationToken,
		Prefix:            &req.Prefix,
	}

	// List objects using the resource manager
	result, err := uc.manager.ListObjects(ctx, provider.ProviderName(req.Provider), req.Definition, listReq)
	if err != nil {
		return nil, fmt.Errorf("failed to list objects: %w", err)
	}

	// Convert to response DTO using factory function
	return dto.NewFileListResponseFromProvider(result, int(req.MaxKeys)), nil
}

// GenerateUploadURL generates a signed URL for file upload
func (uc *FileOperationsUseCase) GenerateUploadURL(ctx context.Context, req *dto.GenerateUploadURLRequest) (*dto.SignedURLResponse, error) {
	opts := req.Upload.To()
	signedURL, err := uc.manager.DefinitionResolver().ResolveUploadURL(ctx, resolver.DefinitionName(req.Definition), opts)
	if err != nil {
		return nil, fmt.Errorf("failed to generate upload URL: %w", err)
	}

	return dto.NewSignedURLResponseFromProvider(signedURL), nil
}

// GenerateDownloadURL generates a signed URL for file download
func (uc *FileOperationsUseCase) GenerateDownloadURL(ctx context.Context, req *dto.GenerateDownloadURLRequest) (*dto.SignedURLResponse, error) {
	var opts *resolver.DownloadOptions
	if req.Download != nil {
		opts = req.Download.To()
	}
	signedURL, err := uc.manager.URLResolver().ResolveDownloadURL(ctx, req.FilePath, opts)
	if err != nil {
		return nil, fmt.Errorf("failed to generate download URL: %w", err)
	}

	return dto.NewSignedURLResponseFromProvider(signedURL), nil
}

// DeleteFile deletes a file from storage
func (uc *FileOperationsUseCase) DeleteFile(ctx context.Context, req *dto.DeleteFileRequest) error {
	err := uc.manager.DeleteObject(ctx, provider.ProviderName(req.Provider), req.FilePath)
	if err != nil {
		return fmt.Errorf("failed to delete file: %w", err)
	}
	return nil
}

// GetFileMetadata retrieves metadata for a specific file
func (uc *FileOperationsUseCase) GetFileMetadata(ctx context.Context, req *dto.GetFileMetadataRequest) (*dto.FileMetadata, error) {
	metadata, err := uc.manager.GetObjectMetadata(ctx, provider.ProviderName(req.Provider), req.FilePath)
	if err != nil {
		return nil, fmt.Errorf("failed to get file metadata: %w", err)
	}

	return dto.NewFileMetadataFromProvider(metadata), nil
}

// UpdateFileMetadata updates metadata for an existing file
func (uc *FileOperationsUseCase) UpdateFileMetadata(ctx context.Context, req *dto.UpdateFileMetadataRequest) (*dto.FileMetadata, error) {
	updateOpts := req.Metadata.ToUpdateMetadata()
	err := uc.manager.UpdateObjectMetadata(ctx, provider.ProviderName(req.Provider), req.FilePath, updateOpts)
	if err != nil {
		return nil, fmt.Errorf("failed to update file metadata: %w", err)
	}

	// Create request for getting metadata
	getReq := &dto.GetFileMetadataRequest{
		Provider: req.Provider,
		FilePath: req.FilePath,
	}
	return uc.GetFileMetadata(ctx, getReq)
}
