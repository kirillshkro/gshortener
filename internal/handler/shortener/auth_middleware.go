package shortener

import (
	"context"
	"net/http"

	"github.com/kirillshkro/gshortener/internal/handler/shortener/claims"
	"github.com/kirillshkro/gshortener/internal/types"
)

// AuthMiddleware is a middleware function that checks if a user is authenticated
// and adds the user ID to the request context if authentication is successful.
// It verifies the authentication cookie and validates the token contained within.
func (s Service) AuthMiddleware(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		// Проверка авторизации пользователя
		userCookie, err := r.Cookie("auth_cookie")
		if err != nil || userCookie.Value == "" {
			s.refreshUserCookie(w)
			next.ServeHTTP(w, r)
			return
		}
		// Если пользователь авторизован, продолжаем выполнуемк запроса
		//извлекаем токен из куки и проверяем его
		token := userCookie.Value
		userID, err := claims.GetUserID(token)
		if err != nil {
			w.WriteHeader(http.StatusUnauthorized)
			next.ServeHTTP(w, r)
			return
		}
		// Если токен валиден, продолжаем
		//добавляем ID пользователя к контексту запроса
		ctx := context.WithValue(r.Context(), types.UserID, userID)
		r = r.WithContext(ctx)
		next.ServeHTTP(w, r)
	}
	return http.HandlerFunc(fn)
}
