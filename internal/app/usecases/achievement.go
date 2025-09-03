package usecases

import (
	"fmt"
	"time"

	"avironactive.com/common/context"
	"avironactive.com/resource"
	"avironactive.com/resource/provider"
	"avironactive.com/resource/resolver"
	"avironactive.com/resource/upload"
	"github.com/google/uuid"

	"github.com/anh-nguyen/resource-server/internal/app/dto"
	"github.com/anh-nguyen/resource-server/internal/domain/entity"
	"github.com/anh-nguyen/resource-server/internal/domain/repository"
)

type AchievementUseCase struct {
	achievementRepo repository.AchievementRepository
	uploadManager   upload.UploadManager
	resourceManager resource.ResourceManager
}

func NewAchievementUseCase(
	achievementRepo repository.AchievementRepository,
	resourceManager resource.ResourceManager,
) *AchievementUseCase {
	return &AchievementUseCase{
		achievementRepo: achievementRepo,
		uploadManager:   resourceManager.UploadManager(),
		resourceManager: resourceManager,
	}
}

func (uc *AchievementUseCase) CreateAchievement(ctx context.Context, req *dto.CreateAchievementRequest) (*dto.CreateAchievementResponse, error) {
	achievement := entity.NewAchievement(req.Name, req.Description)
	achievement.Category = req.Category
	achievement.Points = req.Points

	var uploadResponse *dto.UploadInfo
	var iconURL string
	if req.IconFormat != "" {
		opts := (&resolver.PathDefDownloadOpts{}).
			WithProvider(provider.ProviderName(req.Provider)).
			WithScope(resolver.ScopeGlobal, 0).
			WithValues(map[resolver.ParameterName]string{
				"achievement_id": achievement.ID.String(),
				"format":         req.IconFormat,
			})

		pathResult, err := uc.resourceManager.PathDefinitionResolver().ResolveDownloadURL(
			ctx,
			resolver.PathDefinitionName("achievement"),
			opts,
		)
		if err != nil {
			return nil, fmt.Errorf("failed to resolve achievement path: %w", err)
		}

		iconPath := pathResult.ResolvedPath.Path
		achievement.SetIconPath(iconPath)

		iconURL = pathResult.ObjectURL.URL
		if err := uc.achievementRepo.Create(ctx.Context(), achievement); err != nil {
			return nil, fmt.Errorf("failed to create achievement: %w", err)
		}

		uploadOpts := &upload.UploadOptions{
			ResourceType:     "achievement",
			ResourceID:       achievement.ID.String(),
			ResourceField:    "icon_path",
			ResourceValue:    iconPath,
			ResourceProvider: upload.ResourceProvider(req.Provider),
			UploadType:       upload.UploadTypeSimple,
			PathDefinition:   "achievement",
			StorageProvider:  upload.ResourceProvider(req.Provider),
		}

		pathParams := map[string]string{
			"achievement_id": achievement.ID.String(),
			"format":         req.IconFormat,
		}
		uploadOpts.WithPathParameters(pathParams)

		uploadRecord, err := uc.uploadManager.InitiateUpload(ctx, uploadOpts)
		if err != nil {
			return nil, fmt.Errorf("failed to initiate upload: %w", err)
		}

		signedURL, err := uc.uploadManager.GetSimpleUploadURL(ctx, uploadRecord)
		if err != nil {
			return nil, fmt.Errorf("failed to get upload URL: %w", err)
		}

		uploadResponse = &dto.UploadInfo{
			UploadID:  uploadRecord.ID.String(),
			UploadURL: signedURL.URL,
			ExpiresAt: uploadRecord.ExpiresTime.Unix(),
		}
	} else {
		if err := uc.achievementRepo.Create(ctx.Context(), achievement); err != nil {
			return nil, fmt.Errorf("failed to create achievement: %w", err)
		}
	}

	return &dto.CreateAchievementResponse{
		ID:          achievement.ID.String(),
		Name:        achievement.Name,
		Description: achievement.Description,
		Category:    achievement.Category,
		Points:      achievement.Points,
		IconURL:     iconURL,
		Upload:      uploadResponse,
	}, nil
}

func (uc *AchievementUseCase) GetAchievement(ctx context.Context, id string) (*dto.AchievementResponse, error) {
	achievementID, err := uuid.Parse(id)
	if err != nil {
		return nil, fmt.Errorf("invalid achievement ID: %w", err)
	}

	achievement, err := uc.achievementRepo.GetByID(ctx.Context(), achievementID)
	if err != nil {
		return nil, fmt.Errorf("failed to get achievement: %w", err)
	}

	response := dto.NewAchievementResponse(achievement)
	if achievement.IconPath != "" {
		resolved, err := uc.resourceManager.PathURLResolver().ResolveDownloadURL(ctx, achievement.IconPath, nil)
		if err != nil {
			return nil, err
		}

		response.IconURL = resolved.ObjectURL.URL
	}
	if achievement.BannerPath != "" {
		resolved, err := uc.resourceManager.PathURLResolver().ResolveDownloadURL(ctx, achievement.BannerPath, nil)
		if err != nil {
			return nil, err
		}

		response.BannerURL = resolved.ObjectURL.URL
	}

	return response, nil
}

func (uc *AchievementUseCase) UpdateAchievementIcon(ctx context.Context, req *dto.UpdateIconRequest) (*dto.UpdateIconResponse, error) {
	achievementID, err := uuid.Parse(req.AchievementID)
	if err != nil {
		return nil, fmt.Errorf("invalid achievement ID: %w", err)
	}

	achievement, err := uc.achievementRepo.GetByID(ctx.Context(), achievementID)
	if err != nil {
		return nil, fmt.Errorf("failed to get achievement: %w", err)
	}

	oldIconPath := achievement.IconPath

	opts := (&resolver.PathDefDownloadOpts{}).
		WithProvider(provider.ProviderName(req.Provider)).
		WithScope(resolver.ScopeGlobal, 0).
		WithValues(map[resolver.ParameterName]string{
			"achievement_id": achievementID.String(),
			"format":         req.Format,
		})

	pathResult, err := uc.resourceManager.PathDefinitionResolver().ResolveDownloadURL(
		ctx,
		resolver.PathDefinitionName("achievement"),
		opts,
	)
	if err != nil {
		return nil, fmt.Errorf("failed to resolve achievement path: %w", err)
	}

	if pathResult.ResolvedPath.Path == "" {
		return nil, fmt.Errorf("resolved path is empty")
	}

	newIconPath := pathResult.ResolvedPath.Path
	var uploadRecord *upload.Upload

	if newIconPath != oldIconPath {
		achievement.IconPath = newIconPath
		achievement.UpdatedAt = time.Now()
		if err := uc.achievementRepo.Update(ctx.Context(), achievement); err != nil {
			return nil, fmt.Errorf("failed to update achievement: %w", err)
		}

		uploadOpts := &upload.UploadOptions{
			ResourceType:     "achievement",
			ResourceID:       achievementID.String(),
			ResourceField:    "icon_path",
			ResourceValue:    newIconPath,
			ResourceProvider: upload.ResourceProvider(req.Provider),
			UploadType:       upload.UploadTypeSimple,
			PathDefinition:   "achievement",
			StorageProvider:  upload.ResourceProvider(req.Provider),
		}

		pathParams := map[string]string{
			"achievement_id": achievementID.String(),
			"format":         req.Format,
		}
		uploadOpts.WithPathParameters(pathParams)

		var err error
		uploadRecord, err = uc.uploadManager.InitiateUpload(ctx, uploadOpts)
		if err != nil {
			return nil, fmt.Errorf("failed to initiate upload: %w", err)
		}

		if oldIconPath != "" && oldIconPath != newIconPath {
			go uc.cleanupOldFile(ctx, oldIconPath, req.Provider)
		}
	} else {
		uploadOpts := &upload.UploadOptions{
			ResourceType:     "achievement",
			ResourceID:       achievementID.String(),
			ResourceField:    "icon_path",
			ResourceValue:    newIconPath,
			ResourceProvider: upload.ResourceProvider(req.Provider),
			UploadType:       upload.UploadTypeSimple,
			PathDefinition:   "achievement",
			StorageProvider:  upload.ResourceProvider(req.Provider),
		}

		pathParams := map[string]string{
			"achievement_id": achievementID.String(),
			"format":         req.Format,
		}
		uploadOpts.WithPathParameters(pathParams)

		var err error
		uploadRecord, err = uc.uploadManager.InitiateUpload(ctx, uploadOpts)
		if err != nil {
			return nil, fmt.Errorf("failed to initiate upload: %w", err)
		}
	}

	signedURL, err := uc.uploadManager.GetSimpleUploadURL(ctx, uploadRecord)
	if err != nil {
		return nil, fmt.Errorf("failed to get upload URL: %w", err)
	}

	newIconResolved, err := uc.resourceManager.PathURLResolver().ResolveDownloadURL(ctx, newIconPath, nil)
	if err != nil {
		return nil, err
	}

	return &dto.UpdateIconResponse{
		UploadID:   uploadRecord.ID.String(),
		UploadURL:  signedURL.URL,
		ExpiresAt:  uploadRecord.ExpiresTime.Unix(),
		NewIconURL: newIconResolved.ResolvedPath.Path,
	}, nil
}

func (uc *AchievementUseCase) ConfirmUpload(ctx context.Context, req *dto.ConfirmUploadRequest) error {
	uploadID, err := uuid.Parse(req.UploadID)
	if err != nil {
		return fmt.Errorf("invalid upload ID: %w", err)
	}

	confirmation := &upload.UploadConfirmation{
		Success:  req.Success,
		Error:    req.ErrorMsg,
		FileSize: req.FileSize,
		Metadata: req.Metadata,
	}

	if req.ETags != nil {
		if etags, ok := req.ETags.([]dto.PartETag); ok {
			uploadETags := make([]upload.PartETag, len(etags))
			for i, etag := range etags {
				uploadETags[i] = upload.PartETag{
					Part: etag.Part,
					ETag: etag.ETag,
					Size: etag.Size,
				}
			}
			confirmation.ETags = uploadETags
		} else if etag, ok := req.ETags.(string); ok {
			confirmation.ETags = []upload.PartETag{{Part: 1, ETag: etag}}
		}
	}

	return uc.uploadManager.ConfirmSimpleUpload(ctx, upload.UploadID(uploadID), confirmation)
}

func (uc *AchievementUseCase) ListAchievements(ctx context.Context, offset, limit int) ([]*dto.AchievementResponse, error) {
	achievements, err := uc.achievementRepo.List(ctx.Context(), offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list achievements: %w", err)
	}

	result := make([]*dto.AchievementResponse, len(achievements))
	for i, achievement := range achievements {
		result[i] = dto.NewAchievementResponse(achievement)
		if achievement.IconPath != "" {
			resolved, err := uc.resourceManager.PathURLResolver().ResolveDownloadURL(ctx, achievement.IconPath, nil)
			if err != nil {
				return nil, err
			}

			result[i].IconURL = resolved.ObjectURL.URL
		}
		if achievement.BannerPath != "" {
			resolved, err := uc.resourceManager.PathURLResolver().ResolveDownloadURL(ctx, achievement.BannerPath, nil)
			if err != nil {
				return nil, err
			}

			result[i].BannerURL = resolved.ObjectURL.URL
		}
	}

	return result, nil
}

func (uc *AchievementUseCase) ListActiveAchievements(ctx context.Context, offset, limit int) ([]*dto.AchievementResponse, error) {
	achievements, err := uc.achievementRepo.ListActive(ctx.Context(), offset, limit)
	if err != nil {
		return nil, fmt.Errorf("failed to list active achievements: %w", err)
	}

	result := make([]*dto.AchievementResponse, len(achievements))
	for i, achievement := range achievements {
		result[i] = dto.NewAchievementResponse(achievement)
		if achievement.IconPath != "" {
			resolved, err := uc.resourceManager.PathURLResolver().ResolveDownloadURL(ctx, achievement.IconPath, nil)
			if err != nil {
				return nil, err
			}

			result[i].IconURL = resolved.ObjectURL.URL
		}
		if achievement.BannerPath != "" {
			resolved, err := uc.resourceManager.PathURLResolver().ResolveDownloadURL(ctx, achievement.BannerPath, nil)
			if err != nil {
				return nil, err
			}

			result[i].BannerURL = resolved.ObjectURL.URL
		}
	}

	return result, nil
}

func (uc *AchievementUseCase) cleanupOldFile(ctx context.Context, path, providerName string) {
	err := uc.resourceManager.DeleteObject(ctx, provider.ProviderName(providerName), path)
	if err != nil {
		// Log error but don't fail the main operation
		fmt.Printf("Failed to delete old file %s: %v\n", path, err)
	}
}
