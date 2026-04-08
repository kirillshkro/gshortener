package middleware

import (
	"net/http"
	"time"

	"github.com/kirillshkro/gshortener/internal/config/auth"
	"github.com/kirillshkro/gshortener/internal/handler/shortener/claims"
)

type authWriter struct {
	http.ResponseWriter
	Cookie http.Cookie
}

func (w *authWriter) Write(b []byte) (int, error) {
	return w.ResponseWriter.Write(b)
}

func newAuthWriter(resp http.ResponseWriter) *authWriter {
	cfg := auth.NewAuthConfig()
	authUser := claims.NewAuthUser(cfg)
	token, _ := authUser.Token()
	defaultCookie := http.Cookie{
		Name:     "auth_cookie",
		Value:    token,
		Path:     "/",
		Expires:  time.Now().Add(7 * 24 * time.Hour), // 7 дней
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
		Secure:   false,
	}
	http.SetCookie(resp, &defaultCookie)
	return &authWriter{
		resp,
		defaultCookie,
	}
}

func AuthMiddleware(next http.Handler) http.Handler {
	fn := func(resp http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodPost {
			writer := newAuthWriter(resp)
			next.ServeHTTP(writer, req)
			return
		}
		next.ServeHTTP(resp, req)
	}

	return http.HandlerFunc(fn)
}

// выдает куку
func CookieAuthMiddleware(next http.Handler) http.Handler {
	fn := func(resp http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodPost {
		}
		next.ServeHTTP(resp, req)
	}

	return http.HandlerFunc(fn)
}
