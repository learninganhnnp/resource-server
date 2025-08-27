package handlers

import (
	"strings"

	"github.com/anh-nguyen/resource-server/internal/app/dto"
	"github.com/anh-nguyen/resource-server/internal/app/usecases"
	"github.com/anh-nguyen/resource-server/internal/app/validation"
	"github.com/gofiber/fiber/v2"
)

// FileOperationsHandler handles file operation endpoints
type FileOperationsHandler struct {
	useCase *usecases.FileOperationsUseCase
}

// NewFileOperationsHandler creates a new file operations handler
func NewFileOperationsHandler(useCase *usecases.FileOperationsUseCase) *FileOperationsHandler {
	return &FileOperationsHandler{
		useCase: useCase,
	}
}

// ListFiles handles GET /api/v1/resources/:provider/:definition
func (h *FileOperationsHandler) ListFiles(c *fiber.Ctx) error {
	provider := c.Params("provider")
	definition := c.Params("definition")

	// Validate path parameters
	if err := validation.ValidateProvider(provider); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.ErrorResponse("INVALID_PROVIDER", "Invalid provider", err.Error()),
		)
	}

	if err := validation.ValidateDefinition(definition); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.ErrorResponse("INVALID_DEFINITION", "Invalid definition", err.Error()),
		)
	}

	// Parse query parameters
	maxKeys := c.QueryInt("max_keys", 1000)
	continuationToken := c.Query("continuation_token")
	prefix := c.Query("prefix")

	// Validate query parameters
	if validationErrors := validation.ValidateListParameters(maxKeys, continuationToken, prefix); len(validationErrors) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.ErrorResponse("VALIDATION_ERROR", "Invalid query parameters", validationErrors.Error()),
		)
	}

	// Call use case
	result, err := h.useCase.ListFiles(toContext(c), provider, definition, int32(maxKeys), continuationToken, prefix)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			dto.ErrorResponse("LIST_FILES_ERROR", "Failed to list files", err.Error()),
		)
	}

	return c.JSON(dto.SuccessResponse(result))
}

// GenerateUploadURL handles POST /api/v1/resources/:provider/:definition/upload
func (h *FileOperationsHandler) GenerateUploadURL(c *fiber.Ctx) error {
	provider := c.Params("provider")
	definition := c.Params("definition")

	// Validate path parameters
	if err := validation.ValidateProvider(provider); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.ErrorResponse("INVALID_PROVIDER", "Invalid provider", err.Error()),
		)
	}

	if err := validation.ValidateDefinition(definition); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.ErrorResponse("INVALID_DEFINITION", "Invalid definition", err.Error()),
		)
	}

	// Parse request body
	var req dto.UploadRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.ErrorResponse("INVALID_REQUEST", "Invalid request body", err.Error()),
		)
	}

	// Validate request body
	if validationErrors := validation.ValidateStruct(&req); len(validationErrors) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.ErrorResponse("VALIDATION_ERROR", "Invalid request data", validationErrors.Error()),
		)
	}

	// Call use case
	result, err := h.useCase.GenerateUploadURL(toContext(c), provider, definition, &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			dto.ErrorResponse("UPLOAD_URL_ERROR", "Failed to generate upload URL", err.Error()),
		)
	}

	return c.JSON(dto.SuccessResponse(result))
}

// GenerateMultipartUploadURLs handles POST /api/v1/resources/:provider/:definition/upload/multipart
func (h *FileOperationsHandler) GenerateMultipartUploadURLs(c *fiber.Ctx) error {
	provider := c.Params("provider")
	definition := c.Params("definition")

	// Validate path parameters
	if err := validation.ValidateProvider(provider); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.ErrorResponse("INVALID_PROVIDER", "Invalid provider", err.Error()),
		)
	}

	if err := validation.ValidateDefinition(definition); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.ErrorResponse("INVALID_DEFINITION", "Invalid definition", err.Error()),
		)
	}

	// Parse request body
	var req dto.MultipartUploadRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.ErrorResponse("INVALID_REQUEST", "Invalid request body", err.Error()),
		)
	}

	// Validate request body
	if validationErrors := validation.ValidateStruct(&req); len(validationErrors) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.ErrorResponse("VALIDATION_ERROR", "Invalid request data", validationErrors.Error()),
		)
	}

	// Call use case
	result, err := h.useCase.GenerateMultipartUploadURLs(toContext(c), provider, definition, &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			dto.ErrorResponse("MULTIPART_ERROR", "Failed to generate multipart URLs", err.Error()),
		)
	}

	return c.JSON(dto.SuccessResponse(result))
}

// GenerateDownloadURL handles POST /api/v1/resources/:provider/*/download
func (h *FileOperationsHandler) GenerateDownloadURL(c *fiber.Ctx) error {
	provider := c.Params("provider")
	// Get the wildcard path parameter
	filePath := c.Params("*")

	// Remove the "/download" suffix from the path
	if strings.HasSuffix(filePath, "/download") {
		filePath = strings.TrimSuffix(filePath, "/download")
	}

	// Validate path parameters
	if err := validation.ValidateProvider(provider); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.ErrorResponse("INVALID_PROVIDER", "Invalid provider", err.Error()),
		)
	}

	if err := validation.ValidateFilePath(filePath); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.ErrorResponse("INVALID_FILE_PATH", "Invalid file path", err.Error()),
		)
	}

	// Parse request body
	var req dto.DownloadRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.ErrorResponse("INVALID_REQUEST", "Invalid request body", err.Error()),
		)
	}

	// Validate request body
	if validationErrors := validation.ValidateStruct(&req); len(validationErrors) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.ErrorResponse("VALIDATION_ERROR", "Invalid request data", validationErrors.Error()),
		)
	}

	// Call use case
	result, err := h.useCase.GenerateDownloadURL(toContext(c), provider, filePath, &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			dto.ErrorResponse("DOWNLOAD_URL_ERROR", "Failed to generate download URL", err.Error()),
		)
	}

	return c.JSON(dto.SuccessResponse(result))
}

// DeleteFile handles DELETE /api/v1/resources/:provider/*
func (h *FileOperationsHandler) DeleteFile(c *fiber.Ctx) error {
	provider := c.Params("provider")
	filePath := c.Params("*")

	// Validate path parameters
	if err := validation.ValidateProvider(provider); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.ErrorResponse("INVALID_PROVIDER", "Invalid provider", err.Error()),
		)
	}

	if err := validation.ValidateFilePath(filePath); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.ErrorResponse("INVALID_FILE_PATH", "Invalid file path", err.Error()),
		)
	}

	// Call use case
	err := h.useCase.DeleteFile(toContext(c), provider, filePath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			dto.ErrorResponse("DELETE_ERROR", "Failed to delete file", err.Error()),
		)
	}

	return c.JSON(dto.SuccessResponseWithMessage(nil, "File deleted successfully"))
}

// GetFileMetadata handles GET /api/v1/resources/:provider/*/metadata
func (h *FileOperationsHandler) GetFileMetadata(c *fiber.Ctx) error {
	provider := c.Params("provider")
	// Get the wildcard path parameter and remove "/metadata" suffix
	filePath := c.Params("*")
	if strings.HasSuffix(filePath, "/metadata") {
		filePath = strings.TrimSuffix(filePath, "/metadata")
	}

	// Validate path parameters
	if err := validation.ValidateProvider(provider); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.ErrorResponse("INVALID_PROVIDER", "Invalid provider", err.Error()),
		)
	}

	if err := validation.ValidateFilePath(filePath); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.ErrorResponse("INVALID_FILE_PATH", "Invalid file path", err.Error()),
		)
	}

	// Call use case
	result, err := h.useCase.GetFileMetadata(toContext(c), provider, filePath)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			dto.ErrorResponse("METADATA_ERROR", "Failed to get file metadata", err.Error()),
		)
	}

	return c.JSON(dto.SuccessResponse(result))
}

// UpdateFileMetadata handles PUT /api/v1/resources/:provider/*/metadata
func (h *FileOperationsHandler) UpdateFileMetadata(c *fiber.Ctx) error {
	provider := c.Params("provider")
	// Get the wildcard path parameter and remove "/metadata" suffix
	filePath := c.Params("*")
	if strings.HasSuffix(filePath, "/metadata") {
		filePath = strings.TrimSuffix(filePath, "/metadata")
	}

	// Validate path parameters
	if err := validation.ValidateProvider(provider); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.ErrorResponse("INVALID_PROVIDER", "Invalid provider", err.Error()),
		)
	}

	if err := validation.ValidateFilePath(filePath); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.ErrorResponse("INVALID_FILE_PATH", "Invalid file path", err.Error()),
		)
	}

	// Parse request body
	var req dto.MetadataUpdateRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.ErrorResponse("INVALID_REQUEST", "Invalid request body", err.Error()),
		)
	}

	// Validate request body
	if validationErrors := validation.ValidateStruct(&req); len(validationErrors) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.ErrorResponse("VALIDATION_ERROR", "Invalid request data", validationErrors.Error()),
		)
	}

	// Call use case
	result, err := h.useCase.UpdateFileMetadata(toContext(c), provider, filePath, &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			dto.ErrorResponse("METADATA_UPDATE_ERROR", "Failed to update file metadata", err.Error()),
		)
	}

	return c.JSON(dto.SuccessResponse(result))
}
