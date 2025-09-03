package database

import (
	"context"
	"database/sql"
	"fmt"

	"github.com/anh-nguyen/resource-server/internal/domain/entity"
	"github.com/anh-nguyen/resource-server/internal/domain/repository"
	"github.com/google/uuid"
	"github.com/jackc/pgx/v5/pgxpool"
)

type achievementRepository struct {
	db *pgxpool.Pool
}

func NewAchievementRepository(db *pgxpool.Pool) repository.AchievementRepository {
	return &achievementRepository{db: db}
}

func (r *achievementRepository) Create(ctx context.Context, achievement *entity.Achievement) error {
	query := `
		INSERT INTO achievements (
			id, name, description, icon_path, banner_path, 
			category, points, is_active, metadata, created_at, updated_at
		) VALUES ($1, $2, $3, $4, $5, $6, $7, $8, $9, $10, $11)`

	_, err := r.db.Exec(ctx, query,
		achievement.ID,
		achievement.Name,
		achievement.Description,
		achievement.IconPath,
		achievement.BannerPath,
		achievement.Category,
		achievement.Points,
		achievement.IsActive,
		achievement.Metadata,
		achievement.CreatedAt,
		achievement.UpdatedAt,
	)

	return err
}

func (r *achievementRepository) GetByID(ctx context.Context, id uuid.UUID) (*entity.Achievement, error) {
	query := `
		SELECT id, name, description, icon_path, banner_path,
		       category, points, is_active, metadata, created_at, updated_at
		FROM achievements
		WHERE id = $1`

	var achievement entity.Achievement
	err := r.db.QueryRow(ctx, query, id).Scan(
		&achievement.ID,
		&achievement.Name,
		&achievement.Description,
		&achievement.IconPath,
		&achievement.BannerPath,
		&achievement.Category,
		&achievement.Points,
		&achievement.IsActive,
		&achievement.Metadata,
		&achievement.CreatedAt,
		&achievement.UpdatedAt,
	)

	if err == sql.ErrNoRows {
		return nil, fmt.Errorf("achievement not found")
	}

	return &achievement, err
}

func (r *achievementRepository) Update(ctx context.Context, achievement *entity.Achievement) error {
	query := `
		UPDATE achievements
		SET name = $2, description = $3, icon_path = $4, banner_path = $5,
		    category = $6, points = $7, is_active = $8, metadata = $9, updated_at = $10
		WHERE id = $1`

	_, err := r.db.Exec(ctx, query,
		achievement.ID,
		achievement.Name,
		achievement.Description,
		achievement.IconPath,
		achievement.BannerPath,
		achievement.Category,
		achievement.Points,
		achievement.IsActive,
		achievement.Metadata,
		achievement.UpdatedAt,
	)

	return err
}

func (r *achievementRepository) Delete(ctx context.Context, id uuid.UUID) error {
	query := `DELETE FROM achievements WHERE id = $1`
	_, err := r.db.Exec(ctx, query, id)
	return err
}

func (r *achievementRepository) List(ctx context.Context, offset, limit int) ([]*entity.Achievement, error) {
	query := `
		SELECT id, name, description, icon_path, banner_path,
		       category, points, is_active, metadata, created_at, updated_at
		FROM achievements
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var achievements []*entity.Achievement
	for rows.Next() {
		var achievement entity.Achievement
		err := rows.Scan(
			&achievement.ID,
			&achievement.Name,
			&achievement.Description,
			&achievement.IconPath,
			&achievement.BannerPath,
			&achievement.Category,
			&achievement.Points,
			&achievement.IsActive,
			&achievement.Metadata,
			&achievement.CreatedAt,
			&achievement.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		achievements = append(achievements, &achievement)
	}

	return achievements, nil
}

func (r *achievementRepository) ListActive(ctx context.Context, offset, limit int) ([]*entity.Achievement, error) {
	query := `
		SELECT id, name, description, icon_path, banner_path,
		       category, points, is_active, metadata, created_at, updated_at
		FROM achievements
		WHERE is_active = true
		ORDER BY created_at DESC
		LIMIT $1 OFFSET $2`

	rows, err := r.db.Query(ctx, query, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var achievements []*entity.Achievement
	for rows.Next() {
		var achievement entity.Achievement
		err := rows.Scan(
			&achievement.ID,
			&achievement.Name,
			&achievement.Description,
			&achievement.IconPath,
			&achievement.BannerPath,
			&achievement.Category,
			&achievement.Points,
			&achievement.IsActive,
			&achievement.Metadata,
			&achievement.CreatedAt,
			&achievement.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		achievements = append(achievements, &achievement)
	}

	return achievements, nil
}

func (r *achievementRepository) ListByCategory(ctx context.Context, category string, offset, limit int) ([]*entity.Achievement, error) {
	query := `
		SELECT id, name, description, icon_path, banner_path,
		       category, points, is_active, metadata, created_at, updated_at
		FROM achievements
		WHERE category = $1
		ORDER BY created_at DESC
		LIMIT $2 OFFSET $3`

	rows, err := r.db.Query(ctx, query, category, limit, offset)
	if err != nil {
		return nil, err
	}
	defer rows.Close()

	var achievements []*entity.Achievement
	for rows.Next() {
		var achievement entity.Achievement
		err := rows.Scan(
			&achievement.ID,
			&achievement.Name,
			&achievement.Description,
			&achievement.IconPath,
			&achievement.BannerPath,
			&achievement.Category,
			&achievement.Points,
			&achievement.IsActive,
			&achievement.Metadata,
			&achievement.CreatedAt,
			&achievement.UpdatedAt,
		)
		if err != nil {
			return nil, err
		}
		achievements = append(achievements, &achievement)
	}

	return achievements, nil
}
