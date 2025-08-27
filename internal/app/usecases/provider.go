package usecases

import (
	"fmt"

	"avironactive.com/common/context"
	"avironactive.com/resource/provider"

	"avironactive.com/resource"
	"github.com/anh-nguyen/resource-server/internal/app/dto"
)

// ProviderUseCase handles provider management operations
type ProviderUseCase struct {
	manager resource.ResourceManager
}

// NewProviderUseCase creates a new provider use case
func NewProviderUseCase(manager resource.ResourceManager) *ProviderUseCase {
	return &ProviderUseCase{
		manager: manager,
	}
}

// ListProviders returns all configured storage providers
func (uc *ProviderUseCase) ListProviders(ctx context.Context) ([]dto.ProviderResponse, error) {
	providers := uc.manager.GetAllProviders()

	responses := make([]dto.ProviderResponse, 0, len(providers))
	for _, provider := range providers {
		response := dto.ProviderResponse{
			Name:         string(provider.Name()),
			Capabilities: getProviderCapabilities(provider),
		}
		responses = append(responses, response)
	}

	return responses, nil
}

// GetProvider returns details about a specific provider
func (uc *ProviderUseCase) GetProvider(ctx context.Context, name string) (*dto.ProviderResponse, error) {
	provider, err := uc.manager.GetProvider(provider.ProviderName(name))
	if err != nil {
		return nil, fmt.Errorf("provider not found: %s", name)
	}

	response := &dto.ProviderResponse{
		Name:         name,
		Capabilities: getProviderCapabilities(provider),
	}

	return response, nil
}

// getProviderCapabilities extracts provider capabilities
func getProviderCapabilities(prov provider.Provider) dto.ProviderCapabilities {
	caps := prov.Capabilities()

	var multipart *dto.MultipartCapabilities
	if caps.Multipart != nil {
		multipart = &dto.MultipartCapabilities{
			MinPartSize: caps.Multipart.MinPartSize,
			MaxPartSize: caps.Multipart.MaxPartSize,
			MaxParts:    caps.Multipart.MaxParts,
		}
	}

	return dto.ProviderCapabilities{
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
