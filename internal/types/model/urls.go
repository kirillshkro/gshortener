package model

import "github.com/kirillshkro/gshortener/internal/types"

type DataURL struct {
	ID            uint           `gorm:"not null;primaryKey"`
	ShortURL      types.ShortURL `json:"short_url" gorm:"type:varchar(10);not null;uniqueIndex:idx_short_url"`
	OriginalURL   types.RawURL   `json:"original_url" gorm:"not null;uniqueIndex:idx_original_url"`
	UserProfileID string         `json:"user_profile_uuid" gorm:"type:varchar(36);not null;index:idx_user_profile"`
	UserProfile   UserProfile    `json:"user_profile" gorm:"foreignKey:UserProfileID;references:Id;constraint:OnDelete:CASCADE"`
}
