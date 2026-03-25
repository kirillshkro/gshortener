package types

// Профиль юзера: ID и список его ссылок, признак авторизации
type UserProfile struct {
	Uuid       string    `json:"uuid" gorm:"primaryKey"`
	Authorized bool      `json:"authorized" gorm:"not null"`
	URLs       []DataURL `json:"urls"`
}
