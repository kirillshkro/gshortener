package auth

import (
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

type AuthConfig struct {
	Secret      string        `env:"SECRET_KEY" env-default:"c2VjcmV0a2V5X1NFQ1JFVF9LRVkK"`
	ExpiresTime time.Duration `env:"EXPIRES_TIME" env-default:"168h"`
}

func NewAuthConfig() *AuthConfig {
	var authConfig AuthConfig
	if err := cleanenv.ReadEnv(&authConfig); err != nil {
		panic(err)
	}
	return &authConfig
}
