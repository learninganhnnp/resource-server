package handlers

import (
	"github.com/anh-nguyen/resource-server/internal/app/dto"
	"github.com/anh-nguyen/resource-server/internal/app/usecases"
	"github.com/anh-nguyen/resource-server/internal/app/validation"
	"github.com/gofiber/fiber/v2"
)

// MultipartHandler handles multipart upload endpoints
type MultipartHandler struct {
	useCase *usecases.MultipartUseCase
}

// NewMultipartHandler creates a new multipart handler
func NewMultipartHandler(useCase *usecases.MultipartUseCase) *MultipartHandler {
	return &MultipartHandler{
		useCase: useCase,
	}
}

// InitMultipartUpload handles POST /api/v1/multipart/init
func (h *MultipartHandler) InitMultipartUpload(c *fiber.Ctx) error {
	// Parse request body
	var req dto.MultipartInitRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.NewErrorResponse("INVALID_REQUEST", "Invalid request body", err.Error()),
		)
	}

	// Validate request body
	if validationErrors := validation.ValidateStruct(&req); len(validationErrors) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.NewErrorResponse("VALIDATION_ERROR", "Invalid request data", validationErrors.Error()),
		)
	}

	// Call use case
	result, err := h.useCase.InitMultipartUpload(toContext(c), &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			dto.NewErrorResponse("MULTIPART_INIT_ERROR", "Failed to initialize multipart upload", err.Error()),
		)
	}

	return c.JSON(dto.NewSuccessResponse(result))
}

// GetMultipartURLs handles POST /api/v1/multipart/urls
func (h *MultipartHandler) GetMultipartURLs(c *fiber.Ctx) error {
	// Parse request body
	var req dto.MultipartURLsRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.NewErrorResponse("INVALID_REQUEST", "Invalid request body", err.Error()),
		)
	}

	// Validate request body
	if validationErrors := validation.ValidateStruct(&req); len(validationErrors) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.NewErrorResponse("VALIDATION_ERROR", "Invalid request data", validationErrors.Error()),
		)
	}

	// Call use case
	result, err := h.useCase.GetMultipartURLs(toContext(c), &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			dto.NewErrorResponse("MULTIPART_URLS_ERROR", "Failed to get multipart URLs", err.Error()),
		)
	}

	return c.JSON(dto.NewSuccessResponse(result))
}