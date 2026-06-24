// Package claims provides functionality for JWT token handling and user authentication claims.
package claims

import (
	"log/slog"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/kirillshkro/gshortener/internal/config/auth"
)

// AuthUser represents the structure of JWT claims for authenticated users.
// It embeds jwt.RegisteredClaims and adds custom user-specific fields.
type AuthUser struct {
	jwt.RegisteredClaims
	UserID string
	Cfg    *auth.AuthConfig
}

var logger *slog.Logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
	Level:     slog.LevelDebug,
	AddSource: true,
}))

// NewAuthUser creates and returns a new AuthUser instance with default claims.
// It initializes the user ID with a UUID and sets the expiration time based on the provided configuration.
//
// Parameters:
//   - cfg: pointer to authentication configuration containing expiration time and secret key
//
// Returns:
//   - pointer to newly created AuthUser instance
func NewAuthUser(cfg *auth.AuthConfig) *AuthUser {
	user := &AuthUser{
		RegisteredClaims: jwt.RegisteredClaims{
			ExpiresAt: jwt.NewNumericDate(time.Now().Add(cfg.ExpiresTime)),
		},
		UserID: uuid.NewString(),
		Cfg:    cfg,
	}
	return user
}

// Token generates a JWT token string based on the AuthUser claims.
// It uses HS256 signing method and the secret key from the configuration.
//
// Returns:
//   - string: generated JWT token
//   - error: if token generation fails
func (a AuthUser) Token() (string, error) {
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, a).SignedString([]byte(a.Cfg.Secret))
	if err != nil {
		logger.Error("error creating token", "error", err)
		return "", err
	}
	return token, nil
}

// GetUserID extracts and returns the user ID from a JWT token string.
// If the token is empty, it returns an empty string without error.
//
// Parameters:
//   - token: string representation of JWT token
//
// Returns:
//   - string: extracted user ID
//   - error: if token parsing fails or token is invalid
func GetUserID(token string) (string, error) {
	if token == "" {
		return "", nil
	}
	claims := &AuthUser{}
	if _, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(claims.Cfg.Secret), nil
	}); err != nil {
		return "", err
	}
	return claims.UserID, nil
}
