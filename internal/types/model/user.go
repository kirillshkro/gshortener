package model

import (
	"time"
)

// Профиль юзера: ID и список его ссылок, признак авторизации
type UserProfile struct {
	ID         int       `json:"id" gorm:"primaryKey"`
	Authorized bool      `json:"authorized" gorm:"not null"`
	CreatedAt  time.Time `json:"created_at" gorm:"not null"`
	URLs       []DataURL `json:"urls" gorm:"foreignKey:UserProfileID;references:ID"`
}
