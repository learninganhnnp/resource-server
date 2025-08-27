package handlers

import (
	"avironactive.com/common/context"

	"github.com/anh-nguyen/resource-server/internal/app/dto"
	"github.com/anh-nguyen/resource-server/internal/app/usecases"
	"github.com/gofiber/fiber/v2"
)

// ResourceDefinitionHandler handles resource definition endpoints
type ResourceDefinitionHandler struct {
	useCase *usecases.ResourceDefinitionUseCase
}

// NewResourceDefinitionHandler creates a new resource definition handler
func NewResourceDefinitionHandler(useCase *usecases.ResourceDefinitionUseCase) *ResourceDefinitionHandler {
	return &ResourceDefinitionHandler{
		useCase: useCase,
	}
}

// ListDefinitions handles GET /api/v1/resources/definitions
func (h *ResourceDefinitionHandler) ListDefinitions(c *fiber.Ctx) error {
	definitions, err := h.useCase.ListDefinitions(toContext(c))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			dto.NewErrorResponse("DEFINITIONS_ERROR", "Failed to retrieve definitions", err.Error()),
		)
	}

	return c.JSON(dto.NewSuccessResponse(definitions))
}

// GetDefinition handles GET /api/v1/resources/definitions/:name
func (h *ResourceDefinitionHandler) GetDefinition(c *fiber.Ctx) error {
	name := c.Params("name")
	if name == "" {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.NewErrorResponse("INVALID_PARAMETER", "Definition name is required", ""),
		)
	}

	definition, err := h.useCase.GetDefinition(toContext(c), name)
	if err != nil {
		return c.Status(fiber.StatusNotFound).JSON(
			dto.NewErrorResponse("DEFINITION_NOT_FOUND", "Definition not found", err.Error()),
		)
	}

	return c.JSON(dto.NewSuccessResponse(definition))
}

func toContext(c *fiber.Ctx) context.Context {
	ctx := context.NewContext(c.Context())
	// Add any request-specific values if needed
	ctx.Set("request_id", c.Get("X-Request-ID"))

	return ctx
}
