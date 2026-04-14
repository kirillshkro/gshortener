package shortener

import (
	"encoding/json"
	"net/http"

	"github.com/kirillshkro/gshortener/internal/config/auth"
	"github.com/kirillshkro/gshortener/internal/handler/shortener/claims"
	"github.com/kirillshkro/gshortener/internal/types"
)

type Deleter interface {
	DeleteUserURLs(resp http.ResponseWriter, req *http.Request)
}

func (s Service) DeleteUserURLs(resp http.ResponseWriter, req *http.Request) {

	if !cookieExist(req, "auth_cookie") {
		s.logger.Error("auth cookie not found")
		resp.WriteHeader(http.StatusUnauthorized)
		return
	}

	authUser := claims.NewAuthUser(auth.NewAuthConfig())
	userToken, err := authUser.Token()
	if err != nil {
		s.logger.Error("failed to get user token", "error", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	userID, err := claims.GetUserID(userToken)
	if err != nil {
		s.logger.Error("failed to get user ID", "error", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	var urls []types.ShortURL

	if err := json.NewDecoder(req.Body).Decode(&urls); err != nil {
		s.logger.Error("failed to decode request body", "error", err)
		return
	}

	for _, url := range urls {
		go s.Stor.DeleteUserURL(userID, url)
		s.logger.Info("deleted user URL", "url", url)
	}

	resp.WriteHeader(http.StatusAccepted)
}
