package shortener

import (
	"io"
	"net/http"

	"github.com/kirillshkro/gshortener/internal/config/auth"
	"github.com/kirillshkro/gshortener/internal/handler/shortener/claims"
)

func cookieExist(req *http.Request, cookieName string) bool {
	newReq := io.NopCloser(req.Body)
	_, err := req.Cookie(cookieName)
	if err != nil {
		return false
	}
	req.Body = newReq
	return true
}

func (s Service) createCookie(resp http.ResponseWriter) {
	token, err := s.generateAuthToken()
	if err != nil {
		s.logger.Error("cannot generate token: ", "error: ", err.Error())
		resp.WriteHeader(http.StatusNoContent)
		return
	}
	cookie := &http.Cookie{
		Name:     "auth_cookie",
		Value:    token,
		MaxAge:   7 * 24 * 3600,
		Path:     "/",
		HttpOnly: true,
		Secure:   false,
	}
	http.SetCookie(resp, cookie)
}

func (s Service) refreshUserCookie(resp http.ResponseWriter) {
	token, err := s.generateAuthToken()
	if err != nil {
		s.logger.Error("cannot refresh token: ", "error: ", err.Error())
		return
	}

	http.SetCookie(resp, &http.Cookie{
		Name:     "auth_cookie",
		Value:    token,
		MaxAge:   3600 * 24 * 7,
		Path:     "/",
		Secure:   false,
		HttpOnly: true,
	})
}

func (s Service) generateAuthToken() (string, error) {
	authCfg := auth.NewAuthConfig()
	authUser := claims.NewAuthUser(authCfg)
	return authUser.Token()
}
