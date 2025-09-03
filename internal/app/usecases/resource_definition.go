package usecases

import (
	"avironactive.com/common/context"

	"avironactive.com/resource"
	"avironactive.com/resource/resolver"
	"github.com/anh-nguyen/resource-server/internal/app/dto"
)

// ResourceDefinitionUseCase handles resource definition operations
type ResourceDefinitionUseCase struct {
	manager resource.ResourceManager
}

// NewResourceDefinitionUseCase creates a new resource definition use case
func NewResourceDefinitionUseCase(manager resource.ResourceManager) *ResourceDefinitionUseCase {
	return &ResourceDefinitionUseCase{
		manager: manager,
	}
}

// ListDefinitions returns all available resource path definitions
func (uc *ResourceDefinitionUseCase) ListDefinitions(ctx context.Context) ([]*dto.PathDefinitionResponse, error) {
	definitions := uc.manager.GetAllDefinitions()

	responses := make([]*dto.PathDefinitionResponse, 0, len(definitions))
	for _, def := range definitions {
		responses = append(responses, convertDefinitionPath(def))
	}

	return responses, nil
}

// GetDefinition returns a specific resource definition by name
func (uc *ResourceDefinitionUseCase) GetDefinition(ctx context.Context, name string) (*dto.PathDefinitionResponse, error) {
	def, err := uc.manager.GetDefinition(resolver.PathDefinitionName(name))
	if err != nil {
		return nil, err
	}

	return convertDefinitionPath(def), nil
}

// convertScopes converts resource scopes to string representations
func convertScopes(scopes []resolver.ScopeType) []string {
	result := make([]string, 0, len(scopes))
	for _, scope := range scopes {
		switch scope {
		case resolver.ScopeGlobal:
			result = append(result, "G")
		case resolver.ScopeApp:
			result = append(result, "A")
		case resolver.ScopeClientApp:
			result = append(result, "CA")
		}
	}
	return result
}

// convertParameters converts path parameters to response DTOs
func convertParameters(params []*resolver.ParameterDefinition) []dto.PathParameterResponse {
	result := make([]dto.PathParameterResponse, 0, len(params))
	for _, param := range params {
		response := dto.PathParameterResponse{
			Name: string(param.Name),
			//Required:     param.Required,
			Description:  param.Description,
			DefaultValue: param.DefaultValue,
		}
		result = append(result, response)
	}
	return result
}

func convertDefinitionPath(def *resolver.PathDefinition) *dto.PathDefinitionResponse {
	providers := make([]string, 0, len(def.Patterns))
	for p := range def.Patterns {
		providers = append(providers, string(p))
	}

	return &dto.PathDefinitionResponse{
		Name:          string(def.Name),
		DisplayName:   def.DisplayName,
		Description:   def.Description,
		AllowedScopes: convertScopes(def.AllowedScopes),
		Parameters:    convertParameters(def.Parameters),
		Providers:     providers,
	}
}
