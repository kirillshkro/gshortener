package auth

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type AuthConfig struct {
	Secret      string        `env:"SECRET_KEY"`
	ExpiresTime time.Duration `env:"EXPIRES_TIME"`
}

func NewAuthConfig() *AuthConfig {
	var authConfig AuthConfig
	if err := cleanenv.ReadEnv(&authConfig); err != nil {
		panic(err)
	}
	return &authConfig
}
