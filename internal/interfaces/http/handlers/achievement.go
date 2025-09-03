package handlers

import (
	"github.com/gofiber/fiber/v2"
	"github.com/google/uuid"

	"avironactive.com/common/context"
	"avironactive.com/resource/upload"
	"github.com/anh-nguyen/resource-server/internal/app/dto"
	"github.com/anh-nguyen/resource-server/internal/app/usecases"
	"github.com/anh-nguyen/resource-server/internal/app/validation"
)

type AchievementHandler struct {
	useCase       *usecases.AchievementUseCase
	uploadManager upload.UploadManager
}

func NewAchievementHandler(
	useCase *usecases.AchievementUseCase,
	uploadManager upload.UploadManager,
) *AchievementHandler {
	return &AchievementHandler{
		useCase:       useCase,
		uploadManager: uploadManager,
	}
}

func (h *AchievementHandler) CreateAchievement(c *fiber.Ctx) error {
	var req dto.CreateAchievementRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.NewErrorResponse("INVALID_REQUEST", "Invalid request body", err.Error()),
		)
	}

	if validationErrors := validation.ValidateStruct(&req); len(validationErrors) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.NewErrorResponse("VALIDATION_ERROR", "Invalid request data", validationErrors.Error()),
		)
	}

	if req.Provider == "" {
		req.Provider = "r2"
	}

	ctx := context.Background()
	result, err := h.useCase.CreateAchievement(ctx, &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			dto.NewErrorResponse("CREATE_ERROR", "Failed to create achievement", err.Error()),
		)
	}

	return c.Status(fiber.StatusCreated).JSON(dto.NewSuccessResponse(result))
}

func (h *AchievementHandler) GetAchievement(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := validation.ValidateUUID(id); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.NewErrorResponse("INVALID_ID", "Invalid achievement ID", err.Error()),
		)
	}

	ctx := context.Background()
	result, err := h.useCase.GetAchievement(ctx, id)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			dto.NewErrorResponse("GET_ERROR", "Failed to get achievement", err.Error()),
		)
	}

	return c.JSON(dto.NewSuccessResponse(result))
}

func (h *AchievementHandler) UpdateAchievementIcon(c *fiber.Ctx) error {
	id := c.Params("id")

	if err := validation.ValidateUUID(id); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.NewErrorResponse("INVALID_ID", "Invalid achievement ID", err.Error()),
		)
	}

	var req dto.UpdateIconRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.NewErrorResponse("INVALID_REQUEST", "Invalid request body", err.Error()),
		)
	}

	req.AchievementID = id

	if validationErrors := validation.ValidateStruct(&req); len(validationErrors) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.NewErrorResponse("VALIDATION_ERROR", "Invalid request data", validationErrors.Error()),
		)
	}

	ctx := context.Background()
	result, err := h.useCase.UpdateAchievementIcon(ctx, &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			dto.NewErrorResponse("UPDATE_ERROR", "Failed to update achievement icon", err.Error()),
		)
	}

	return c.JSON(dto.NewSuccessResponse(result))
}

func (h *AchievementHandler) ConfirmUpload(c *fiber.Ctx) error {
	uploadID := c.Params("id")

	if err := validation.ValidateUUID(uploadID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.NewErrorResponse("INVALID_ID", "Invalid upload ID", err.Error()),
		)
	}

	var req dto.ConfirmUploadRequest
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.NewErrorResponse("INVALID_REQUEST", "Invalid request body", err.Error()),
		)
	}

	req.UploadID = uploadID

	if validationErrors := validation.ValidateStruct(&req); len(validationErrors) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.NewErrorResponse("VALIDATION_ERROR", "Invalid request data", validationErrors.Error()),
		)
	}

	ctx := context.Background()
	err := h.useCase.ConfirmUpload(ctx, &req)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			dto.NewErrorResponse("CONFIRM_ERROR", "Failed to confirm upload", err.Error()),
		)
	}

	return c.JSON(dto.NewSuccessResponseWithMessage(nil, "Upload confirmed successfully"))
}

func (h *AchievementHandler) GetMultipartURLs(c *fiber.Ctx) error {
	uploadID := c.Params("id")

	if err := validation.ValidateUUID(uploadID); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.NewErrorResponse("INVALID_ID", "Invalid upload ID", err.Error()),
		)
	}

	var req struct {
		PartCount int `json:"part_count" validate:"required,min=1,max=10000"`
	}
	if err := c.BodyParser(&req); err != nil {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.NewErrorResponse("INVALID_REQUEST", "Invalid request body", err.Error()),
		)
	}

	if validationErrors := validation.ValidateStruct(&req); len(validationErrors) > 0 {
		return c.Status(fiber.StatusBadRequest).JSON(
			dto.NewErrorResponse("VALIDATION_ERROR", "Invalid request data", validationErrors.Error()),
		)
	}

	uploadIDUUID, _ := uuid.Parse(uploadID)
	ctx := context.Background()

	uploadRecord, err := h.uploadManager.GetUpload(ctx, upload.UploadID(uploadIDUUID))
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			dto.NewErrorResponse("UPLOAD_NOT_FOUND", "Upload not found", err.Error()),
		)
	}

	multipartURLs, err := h.uploadManager.GetPartURLs(ctx, uploadRecord, req.PartCount)
	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			dto.NewErrorResponse("MULTIPART_URLS_ERROR", "Failed to get multipart URLs", err.Error()),
		)
	}

	return c.JSON(dto.NewSuccessResponse(multipartURLs))
}

func (h *AchievementHandler) ListAchievements(c *fiber.Ctx) error {
	page := c.QueryInt("page", 1)
	pageSize := c.QueryInt("pageSize", 20)
	onlyActive := c.QueryBool("only_active", true)

	if page < 1 {
		page = 1
	}
	if pageSize < 1 || pageSize > 100 {
		pageSize = 20
	}

	offset := (page - 1) * pageSize

	ctx := context.Background()
	var result []*dto.AchievementResponse
	var err error

	if onlyActive {
		result, err = h.useCase.ListActiveAchievements(ctx, offset, pageSize)
	} else {
		result, err = h.useCase.ListAchievements(ctx, offset, pageSize)
	}

	if err != nil {
		return c.Status(fiber.StatusInternalServerError).JSON(
			dto.NewErrorResponse("LIST_ERROR", "Failed to list achievements", err.Error()),
		)
	}

	response := map[string]interface{}{
		"achievements": result,
		"page":         page,
		"pageSize":     pageSize,
		"total":        len(result),
	}

	return c.JSON(dto.NewSuccessResponse(response))
}
