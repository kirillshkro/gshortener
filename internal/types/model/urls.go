package model

import "github.com/kirillshkro/gshortener/internal/types"

type DataURL struct {
	ID            int            `gorm:"primaryKey"`
	ShortURL      types.ShortURL `json:"short_url" gorm:"type:varchar(10);not null;uniqueIndex"`
	OriginalURL   types.RawURL   `json:"original_url" gorm:"not null;uniqueIndex"`
	UserProfileID int            `json:"user_profile_id" gorm:"not null;index"`
	UserProfile   UserProfile    `gorm:"foreignKey:UserProfileID;references:ID;constraint:OnDelete:CASCADE"`
}
