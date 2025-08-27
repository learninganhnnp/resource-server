package usecases

import (
	"fmt"

	"avironactive.com/common/context"
	"avironactive.com/resource"
	"avironactive.com/resource/provider"
	"github.com/anh-nguyen/resource-server/internal/app/dto"
)

// MultipartUseCase handles multipart upload operations
type MultipartUseCase struct {
	manager resource.ResourceManager
}

// NewMultipartUseCase creates a new multipart use case
func NewMultipartUseCase(manager resource.ResourceManager) *MultipartUseCase {
	return &MultipartUseCase{
		manager: manager,
	}
}

// InitMultipartUpload initializes a multipart upload
func (uc *MultipartUseCase) InitMultipartUpload(ctx context.Context, req *dto.MultipartInitRequest) (*dto.MultipartInitResponse, error) {
	opts := req.To()

	// Get the path using ResolveReadURL
	pathResult, err := uc.manager.PathDefinitionResolver().ResolveDownloadURL(
		ctx,
		resource.PathDefinitionName(req.DefinitionName),
		opts,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve path: %w", err)
	}

	prov, err := uc.manager.GetProvider(provider.ProviderName(req.Provider))
	if err != nil {
		return nil, fmt.Errorf("failed to get provider: %w", err)
	}

	// Check if provider supports multipart
	multipartProvider, ok := prov.(provider.MultipartProvider)
	if !ok {
		return nil, fmt.Errorf("provider %s does not support multipart uploads", req.Provider)
	}

	var headers *provider.RequestHeaders
	if req.Metadata != nil {
		headers = req.Metadata.ToRequestHeaders()
	}

	uploadID, err := multipartProvider.CreateMultipartUpload(ctx.Context(), pathResult.ResolvedPath, headers)
	if err != nil {
		return nil, fmt.Errorf("failed to create multipart upload: %w", err)
	}

	caps := multipartProvider.Capabilities()
	return dto.NewMultipartInitResponse(uploadID, pathResult.ResolvedPath, req.Provider, caps.Multipart), nil
}

// GetMultipartURLs gets signed URLs for multipart upload parts
func (uc *MultipartUseCase) GetMultipartURLs(ctx context.Context, req *dto.MultipartURLsRequest) (*dto.MultipartURLsResponse, error) {
	opts := req.To()
	urlResolver := uc.manager.PathURLResolver()
	multipartResult, err := urlResolver.ResolveMultipartURLs(
		ctx,
		req.Path, // Use the path from the request
		req.UploadID,
		opts,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve multipart URLs: %w", err)
	}

	// Return response using factory function
	return dto.NewMultipartURLsResponse(
		multipartResult.PartURLs,
		&multipartResult.CompleteURL,
		&multipartResult.AbortURL,
	), nil
}
