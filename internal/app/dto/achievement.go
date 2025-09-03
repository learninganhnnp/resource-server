package dto

import (
	"time"

	"avironactive.com/resource/metadata"
	"github.com/anh-nguyen/resource-server/internal/domain/entity"
)

type CreateAchievementRequest struct {
	Name        string `json:"name" validate:"required,max=255"`
	Description string `json:"description" validate:"max=1000"`
	Category    string `json:"category" validate:"omitempty,max=50"`
	Points      int    `json:"points" validate:"min=0,max=10000"`
	IconFormat  string `json:"iconFormat" validate:"omitempty,oneof=png jpg svg webp"`
	Provider    string `json:"provider" validate:"omitempty,oneof=cdn gcs r2"`
}

type CreateAchievementResponse struct {
	ID          string      `json:"id"`
	Name        string      `json:"name"`
	Description string      `json:"description"`
	Category    string      `json:"category,omitempty"`
	Points      int         `json:"points"`
	IconURL     string      `json:"iconUrl,omitempty"`
	Upload      *UploadInfo `json:"upload,omitempty"`
}

type UploadInfo struct {
	UploadID  string `json:"upload_id"`
	UploadURL string `json:"upload_url"`
	ExpiresAt int64  `json:"expires_at"`
}

type AchievementResponse struct {
	ID          string                 `json:"id"`
	Name        string                 `json:"name"`
	Description string                 `json:"description"`
	Category    string                 `json:"category,omitempty"`
	Points      int                    `json:"points"`
	IconURL     string                 `json:"iconUrl,omitempty"`
	BannerURL   string                 `json:"bannerUrl,omitempty"`
	IsActive    bool                   `json:"isActive"`
	Metadata    map[string]interface{} `json:"metadata,omitempty"`
	CreatedAt   time.Time              `json:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt"`
}

type UpdateIconRequest struct {
	AchievementID string `json:"achievement_id" validate:"required,uuid"`
	Format        string `json:"format" validate:"required,oneof=png jpg svg webp"`
	Provider      string `json:"provider" validate:"required,oneof=cdn gcs r2"`
}

type UpdateIconResponse struct {
	UploadID   string `json:"upload_id"`
	UploadURL  string `json:"upload_url"`
	ExpiresAt  int64  `json:"expires_at"`
	NewIconURL string `json:"new_iconUrl"`
}

type PartETag struct {
	Part int    `json:"part"`
	ETag string `json:"etag"`
	Size int64  `json:"size,omitempty"`
}

type ConfirmUploadRequest struct {
	UploadID     string                    `json:"upload_id" validate:"required,uuid"`
	Success      bool                      `json:"success"`
	ErrorMsg     string                    `json:"error_msg,omitempty"`
	FileSize     int64                     `json:"file_size,omitempty"`
	ContentType  string                    `json:"content_type,omitempty"`
	ETags        interface{}               `json:"etags,omitempty"`
	Metadata     *metadata.StorageMetadata `json:"metadata,omitempty"`
	VerifyExists bool                      `json:"verify_exists,omitempty"`
}

type InitMultipartRequest struct {
	Name        string `json:"name" validate:"required,max=255"`
	Description string `json:"description" validate:"max=1000"`
	Category    string `json:"category" validate:"omitempty,max=50"`
	Points      int    `json:"points" validate:"min=0,max=10000"`
	IconFormat  string `json:"iconFormat" validate:"required,oneof=png jpg svg webp"`
	Provider    string `json:"provider" validate:"required,oneof=cdn gcs r2"`
	FileSize    int64  `json:"file_size" validate:"required,min=5242880"`
}

type InitMultipartResponse struct {
	AchievementID     string `json:"achievement_id"`
	UploadID          string `json:"upload_id"`
	MultipartUploadID string `json:"multipart_upload_id"`
	Path              string `json:"path"`
	IconURL           string `json:"iconUrl"`
	MinPartSize       int64  `json:"min_part_size"`
	MaxPartSize       int64  `json:"max_part_size"`
	MaxParts          int    `json:"max_parts"`
}

func NewAchievementResponse(achievement *entity.Achievement) *AchievementResponse {
	return &AchievementResponse{
		ID:          achievement.ID.String(),
		Name:        achievement.Name,
		Description: achievement.Description,
		Category:    achievement.Category,
		Points:      achievement.Points,
		IsActive:    achievement.IsActive,
		Metadata:    achievement.Metadata,
		CreatedAt:   achievement.CreatedAt,
		UpdatedAt:   achievement.UpdatedAt,
	}
}
