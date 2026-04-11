package model

import (
	"github.com/kirillshkro/gshortener/internal/types"
	"gorm.io/gorm"
)

type URLData struct {
	ID          uint           `gorm:"not null;primaryKey"`
	ShortURL    types.ShortURL `json:"short_url" gorm:"not null;uniqueIndex"`
	OriginalURL types.RawURL   `json:"original_url" gorm:"not null;uniqueIndex"`
	UserUUID    string         `json:"user_uuid" gorm:"not null;index"`
	Deleted     gorm.DeletedAt `gorm:"index"`
}
