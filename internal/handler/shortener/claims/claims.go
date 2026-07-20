// Package claims provides a way to handle user authentication using JWT (JSON Web Tokens).
package claims

import (
	"log/slog"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/kirillshkro/gshortener/internal/config/auth"
)

// AuthUser represents a user authenticated in the system. It embeds jwt.RegisteredClaims to provide standard claims,
// such as expiration time and issuer.
type AuthUser struct {
	jwt.RegisteredClaims
	// UserID is a unique identifier for the user.
	UserID string
	// Cfg is the authentication configuration.
	Cfg *auth.AuthConfig
}

var logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
	Level:     slog.LevelDebug,
	AddSource: true,
}))

// NewAuthUser creates a new AuthUser with the given configuration and sets an expiration time based on the
// configuration's ExpiresTime field. The UserID is set to a newly generated UUID.
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

// Token generates a new JWT string for the AuthUser using the secret key from the configuration.
func (a AuthUser) Token() (string, error) {
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, a).SignedString([]byte(a.Cfg.Secret))
	if err != nil {
		logger.Error("error creating token", "error", err)
		return "", err
	}
	return token, nil
}

// GetUserID parses the given JWT string and returns the UserID if it's valid.
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
