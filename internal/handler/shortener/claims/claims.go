package claims

import (
	"log/slog"
	"os"
	"time"

	"github.com/golang-jwt/jwt/v5"
	"github.com/google/uuid"
	"github.com/kirillshkro/gshortener/internal/config/auth"
)

type AuthUser struct {
	jwt.RegisteredClaims
	UserID string
	Cfg    *auth.AuthConfig
}

var logger *slog.Logger = slog.New(slog.NewTextHandler(os.Stdout, &slog.HandlerOptions{
	Level:     slog.LevelDebug,
	AddSource: true,
}))

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

func (a AuthUser) Token() (string, error) {
	token, err := jwt.NewWithClaims(jwt.SigningMethodHS256, a).SignedString([]byte(a.Cfg.Secret))
	if err != nil {
		logger.Error("error creating token", "error", err)
		return "", err
	}
	return token, nil
}

func (a AuthUser) GetUserID(token string) (string, error) {
	if token == "" {
		return "", nil
	}
	claims := &AuthUser{}
	if _, err := jwt.ParseWithClaims(token, claims, func(token *jwt.Token) (interface{}, error) {
		return []byte(a.Cfg.Secret), nil
	}); err != nil {
		return "", err
	}
	return claims.UserID, nil
}
