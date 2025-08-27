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
			dto.NewErrorResponse("INVALID_PROVIDER", "Invalid provider", err.Error()),
		)
	}

	if err := validation.ValidateDefinition(definition); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.NewErrorResponse("INVALID_DEFINITION", "Invalid definition", err.Error()),
		)
	}

	// Parse query parameters
	maxKeys := c.QueryInt("max_keys", 1000)
	continuationToken := c.Query("continuation_token")
	prefix := c.Query("prefix")

	// Validate query parameters
	if validationErrors := validation.ValidateListParameters(maxKeys, continuationToken, prefix); len(validationErrors) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.NewErrorResponse("VALIDATION_ERROR", "Invalid query parameters", validationErrors.Error()),
		)
	}

	// Create structured request
	req := &dto.ListFilesRequest{
		Provider:          provider,
		Definition:        definition,
		MaxKeys:           int32(maxKeys),
		ContinuationToken: continuationToken,
		Prefix:            prefix,
	}

	// Call use case
	result, err := h.useCase.ListFiles(toContext(c), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			dto.NewErrorResponse("LIST_FILES_ERROR", "Failed to list files", err.Error()),
		)
	}

	return c.JSON(dto.NewSuccessResponse(result))
}

// GenerateUploadURL handles POST /api/v1/resources/:provider/:definition/upload
func (h *FileOperationsHandler) GenerateUploadURL(c *fiber.Ctx) error {
	provider := c.Params("provider")
	definition := c.Params("definition")

	// Validate path parameters
	if err := validation.ValidateProvider(provider); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.NewErrorResponse("INVALID_PROVIDER", "Invalid provider", err.Error()),
		)
	}

	if err := validation.ValidateDefinition(definition); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.NewErrorResponse("INVALID_DEFINITION", "Invalid definition", err.Error()),
		)
	}

	// Parse request body
	var req dto.UploadRequest
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

	// Create structured request
	uploadReq := &dto.GenerateUploadURLRequest{
		Provider:   provider,
		Definition: definition,
		Upload:     &req,
	}

	// Call use case
	result, err := h.useCase.GenerateUploadURL(toContext(c), uploadReq)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			dto.NewErrorResponse("UPLOAD_URL_ERROR", "Failed to generate upload URL", err.Error()),
		)
	}

	return c.JSON(dto.NewSuccessResponse(result))
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
			dto.NewErrorResponse("INVALID_PROVIDER", "Invalid provider", err.Error()),
		)
	}

	if err := validation.ValidateFilePath(filePath); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.NewErrorResponse("INVALID_FILE_PATH", "Invalid file path", err.Error()),
		)
	}

	// Parse request body
	var req dto.DownloadRequest
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

	// Create structured request
	downloadReq := &dto.GenerateDownloadURLRequest{
		Provider: provider,
		FilePath: filePath,
		Download: &req,
	}

	// Call use case
	result, err := h.useCase.GenerateDownloadURL(toContext(c), downloadReq)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			dto.NewErrorResponse("DOWNLOAD_URL_ERROR", "Failed to generate download URL", err.Error()),
		)
	}

	return c.JSON(dto.NewSuccessResponse(result))
}

// DeleteFile handles DELETE /api/v1/resources/:provider/*
func (h *FileOperationsHandler) DeleteFile(c *fiber.Ctx) error {
	provider := c.Params("provider")
	filePath := c.Params("*")

	// Validate path parameters
	if err := validation.ValidateProvider(provider); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.NewErrorResponse("INVALID_PROVIDER", "Invalid provider", err.Error()),
		)
	}

	if err := validation.ValidateFilePath(filePath); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.NewErrorResponse("INVALID_FILE_PATH", "Invalid file path", err.Error()),
		)
	}

	// Create structured request
	req := &dto.DeleteFileRequest{
		Provider: provider,
		FilePath: filePath,
	}

	// Call use case
	err := h.useCase.DeleteFile(toContext(c), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			dto.NewErrorResponse("DELETE_ERROR", "Failed to delete file", err.Error()),
		)
	}

	return c.JSON(dto.NewSuccessResponseWithMessage(nil, "File deleted successfully"))
}

// GetFileMetadata handles GET /api/v1/resources/:provider/*/metadata
func (h *FileOperationsHandler) GetFileMetadata(c *fiber.Ctx) error {
	provider := c.Params("provider")
	// Get the wildcard path parameter and remove "/metadata" suffix
	filePath := c.Params("*")
	filePath = strings.TrimSuffix(filePath, "/metadata")

	// Validate path parameters
	if err := validation.ValidateProvider(provider); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.NewErrorResponse("INVALID_PROVIDER", "Invalid provider", err.Error()),
		)
	}

	if err := validation.ValidateFilePath(filePath); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.NewErrorResponse("INVALID_FILE_PATH", "Invalid file path", err.Error()),
		)
	}

	// Create structured request
	req := &dto.GetFileMetadataRequest{
		Provider: provider,
		FilePath: filePath,
	}

	// Call use case
	result, err := h.useCase.GetFileMetadata(toContext(c), req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			dto.NewErrorResponse("METADATA_ERROR", "Failed to get file metadata", err.Error()),
		)
	}

	return c.JSON(dto.NewSuccessResponse(result))
}

// UpdateFileMetadata handles PUT /api/v1/resources/:provider/*/metadata
func (h *FileOperationsHandler) UpdateFileMetadata(c *fiber.Ctx) error {
	provider := c.Params("provider")
	// Get the wildcard path parameter and remove "/metadata" suffix
	filePath := c.Params("*")
	filePath = strings.TrimSuffix(filePath, "/metadata")

	// Validate path parameters
	if err := validation.ValidateProvider(provider); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.NewErrorResponse("INVALID_PROVIDER", "Invalid provider", err.Error()),
		)
	}

	if err := validation.ValidateFilePath(filePath); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.NewErrorResponse("INVALID_FILE_PATH", "Invalid file path", err.Error()),
		)
	}

	// Parse request body
	var req dto.MetadataUpdateRequest
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

	// Create structured request
	updateReq := &dto.UpdateFileMetadataRequest{
		Provider: provider,
		FilePath: filePath,
		Metadata: &req,
	}

	// Call use case
	result, err := h.useCase.UpdateFileMetadata(toContext(c), updateReq)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			dto.NewErrorResponse("METADATA_UPDATE_ERROR", "Failed to update file metadata", err.Error()),
		)
	}

	return c.JSON(dto.NewSuccessResponse(result))
}
