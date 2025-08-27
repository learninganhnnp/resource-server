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
func (uc *ProviderUseCase) ListProviders(ctx context.Context) ([]*dto.ProviderResponse, error) {
	providers := uc.manager.GetAllProviders()

	responses := make([]*dto.ProviderResponse, 0, len(providers))
	for _, provider := range providers {
		responses = append(responses, dto.NewProviderResponse(provider))
	}

	return responses, nil
}

// GetProvider returns details about a specific provider
func (uc *ProviderUseCase) GetProvider(ctx context.Context, name string) (*dto.ProviderResponse, error) {
	provider, err := uc.manager.GetProvider(provider.ProviderName(name))
	if err != nil {
		return nil, fmt.Errorf("provider not found: %s", name)
	}

	return dto.NewProviderResponse(provider), nil
}
