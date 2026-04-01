package model

import (
	"time"
)

// Профиль юзера: ID и список его ссылок, признак авторизации
type UserProfile struct {
	Id         int       `json:"uuid" gorm:"primaryKey"`
	Authorized bool      `json:"authorized" gorm:"not null"`
	CreatedAt  time.Time `json:"created_at" gorm:"not null"`
	URLs       []DataURL `json:"urls" gorm:"foreignKey:UserProfileID;constraint:OnDelete:CASCADE"`
}
