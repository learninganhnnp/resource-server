package entity

import (
	"time"

	"github.com/google/uuid"
)

type Achievement struct {
	ID          uuid.UUID              `json:"id" db:"id"`
	Name        string                 `json:"name" db:"name"`
	Description string                 `json:"description" db:"description"`
	IconPath    string                 `json:"icon_path" db:"icon_path"`
	BannerPath  string                 `json:"banner_path" db:"banner_path"`
	Category    string                 `json:"category" db:"category"`
	Points      int                    `json:"points" db:"points"`
	IsActive    bool                   `json:"is_active" db:"is_active"`
	Metadata    map[string]interface{} `json:"metadata" db:"metadata"`
	CreatedAt   time.Time              `json:"createdAt" db:"createdAt"`
	UpdatedAt   time.Time              `json:"updatedAt" db:"updatedAt"`

	IconURL   string `json:"iconUrl,omitempty" db:"-"`
	BannerURL string `json:"bannerUrl,omitempty" db:"-"`
}

func NewAchievement(name, description string) *Achievement {
	return &Achievement{
		ID:          uuid.New(),
		Name:        name,
		Description: description,
		IsActive:    true,
		CreatedAt:   time.Now(),
		UpdatedAt:   time.Now(),
	}
}

func (a *Achievement) SetIconPath(iconPath string) {
	a.IconPath = iconPath
	a.UpdatedAt = time.Now()
}

func (a *Achievement) SetBannerPath(bannerPath string) {
	a.BannerPath = bannerPath
	a.UpdatedAt = time.Now()
}
