// Package auth provides configuration for the authentication system. It uses environment variables to configure itself,
// with default values provided if no corresponding environment variable is set.
package auth

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

// AuthConfig represents the configuration for the authentication system. It includes fields
// for a secret key and an expiration time for tokens.
type AuthConfig struct {
	// Secret is the secret key used to sign tokens.
	Secret string `env:"SECRET_KEY" env-default:"c2VjcmV0a2V5X1NFQ1JFVF9LRVkK"`
	// ExpiresTime is the duration after which a token should be considered expired.
	ExpiresTime time.Duration `env:"EXPIRES_TIME" env-default:"168h"`
}

// NewAuthConfig returns a new AuthConfig instance, populated with values from environment variables.
// If an error occurs while reading the environment variables, it will be logged and the program will panic.
func NewAuthConfig() *AuthConfig {
	var authConfig AuthConfig
	if err := cleanenv.ReadEnv(&authConfig); err != nil {
		log.Fatal(err)
	}
	return &authConfig
}
