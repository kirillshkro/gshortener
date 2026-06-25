// Package model contains data models for the URL shortener application
package model

import "github.com/kirillshkro/gshortener/internal/types"

// URLData represents the structure of a URL record in the database
// It stores information about short URLs, their original URLs, associated user, and deletion status
type URLData struct {
	// ID is the primary key of the URL record
	ID uint `gorm:"not null;primaryKey"`
	// ShortURL is the shortened version of the original URL
	ShortURL types.ShortURL `json:"short_url" gorm:"not null;uniqueIndex"`
	// OriginalURL is the full, original URL that was shortened
	OriginalURL types.RawURL `json:"original_url" gorm:"not null;uniqueIndex"`
	// UserUUID is the identifier of the user who created this URL
	UserUUID string `json:"user_uuid" gorm:"not null;index"`
	// IsDeleted indicates whether the URL has been marked as deleted
	IsDeleted bool `gorm:"not null;default:false;index"`
}
