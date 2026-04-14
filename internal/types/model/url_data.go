package model

import "github.com/kirillshkro/gshortener/internal/types"

type URLData struct {
	ID          uint           `gorm:"not null;primaryKey"`
	ShortURL    types.ShortURL `json:"short_url" gorm:"not null;uniqueIndex"`
	OriginalURL types.RawURL   `json:"original_url" gorm:"not null;uniqueIndex"`
	UserID      string         `json:"user_uuid" gorm:"not null;index"`
	IsDeleted   bool           `gorm:"default:false"`
}
