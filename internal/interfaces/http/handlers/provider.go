package handlers

import (
	"github.com/anh-nguyen/resource-server/internal/app/dto"
	"github.com/anh-nguyen/resource-server/internal/app/usecases"
	"github.com/gofiber/fiber/v2"
)

// ProviderHandler handles provider management endpoints
type ProviderHandler struct {
	useCase *usecases.ProviderUseCase
}

// NewProviderHandler creates a new provider handler
func NewProviderHandler(useCase *usecases.ProviderUseCase) *ProviderHandler {
	return &ProviderHandler{
		useCase: useCase,
	}
}

// ListProviders handles GET /api/v1/resources/providers
func (h *ProviderHandler) ListProviders(c *fiber.Ctx) error {
	providers, err := h.useCase.ListProviders(toContext(c))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			dto.NewErrorResponse("PROVIDERS_ERROR", "Failed to retrieve providers", err.Error()),
		)
	}

	return c.JSON(dto.NewSuccessResponse(providers))
}

// GetProvider handles GET /api/v1/resources/providers/:name
func (h *ProviderHandler) GetProvider(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.NewErrorResponse("INVALID_PARAMETER", "Provider name is required", ""),
		)
	}

	provider, err := h.useCase.GetProvider(toContext(c), name)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			dto.NewErrorResponse("PROVIDER_NOT_FOUND", "Provider not found", err.Error()),
		)
	}

	return c.JSON(dto.NewSuccessResponse(provider))
}
