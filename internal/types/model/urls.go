package model

import "github.com/kirillshkro/gshortener/internal/types"

type DataURL struct {
	ID            uint           `gorm:"not null;primaryKey"`
	ShortURL      types.ShortURL `json:"short_url" gorm:"type:varchar(10);not null;uniqueIndex"`
	OriginalURL   types.RawURL   `json:"original_url" gorm:"not null;uniqueIndex"`
	UserProfileID int            `json:"user_profile_uuid" gorm:"not null;index"`
	UserProfile   UserProfile    `json:"user_profile" gorm:"foreignKey:UserProfileID;references:Id;constraint:OnDelete:CASCADE"`
}
