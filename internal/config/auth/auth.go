// Package auth provides configuration and management for authentication
package auth

import (
	"log"
	"time"

	"github.com/ilyakaznacheev/cleanenv"
)

// AuthConfig represents the authentication configuration structure
// It holds the secret key and expiration time for authentication tokens
type AuthConfig struct {
	// Secret is the secret key used for signing authentication tokens
	// It should be a secure, random string
	Secret string `env:"SECRET_KEY" env-default:"c2VjcmV0a2V5X1NFQ1JFVF9LRVkK"`
	// ExpiresTime is the duration after which authentication tokens expire
	// Default value is 168 hours (7 days)
	ExpiresTime time.Duration `env:"EXPIRES_TIME" env-default:"168h"`
}

// NewAuthConfig creates and returns a new AuthConfig instance
// It reads configuration values from environment variables
// If environment variables are not set, it uses default values
// In case of error during reading configuration, it logs the error and exits the program
// Returns pointer to AuthConfig
func NewAuthConfig() *AuthConfig {
	var authConfig AuthConfig
	if err := cleanenv.ReadEnv(&authConfig); err != nil {
		log.Fatal(err)
	}
	return &authConfig
}
