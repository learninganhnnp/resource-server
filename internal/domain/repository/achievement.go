package repository

import (
	"context"

	"github.com/google/uuid"
	"github.com/anh-nguyen/resource-server/internal/domain/entity"
)

type AchievementRepository interface {
	Create(ctx context.Context, achievement *entity.Achievement) error
	GetByID(ctx context.Context, id uuid.UUID) (*entity.Achievement, error)
	Update(ctx context.Context, achievement *entity.Achievement) error
	Delete(ctx context.Context, id uuid.UUID) error
	List(ctx context.Context, offset, limit int) ([]*entity.Achievement, error)
	ListActive(ctx context.Context, offset, limit int) ([]*entity.Achievement, error)
	ListByCategory(ctx context.Context, category string, offset, limit int) ([]*entity.Achievement, error)
}