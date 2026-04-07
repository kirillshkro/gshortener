package middleware

import (
	"net/http"
	"time"

	"github.com/google/uuid"
)

type authWriter struct {
	http.ResponseWriter
	Cookie http.Cookie
}

func (w *authWriter) Write(b []byte) (int, error) {
	return w.ResponseWriter.Write(b)
}

func newAuthWriter(resp http.ResponseWriter, cookiePath string) *authWriter {
	defaultCookie := http.Cookie{
		Name:     "auth_cookie",
		Value:    uuid.NewString(),
		Path:     cookiePath,
		Expires:  time.Now().Add(7 * 27 * time.Hour),
		HttpOnly: true,
		SameSite: http.SameSiteStrictMode,
	}
	return &authWriter{
		resp,
		defaultCookie,
	}
}

func AuthMiddleware(next http.Handler) http.Handler {
	fn := func(resp http.ResponseWriter, req *http.Request) {
		if req.Method == http.MethodPost {
			writer := newAuthWriter(resp, req.URL.Path)
			next.ServeHTTP(writer, req)
			return
		}
		next.ServeHTTP(resp, req)
	}

	return http.HandlerFunc(fn)
}
