package shortener

import (
	"encoding/json"
	"net/http"
	"sync"

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

	var (
		urls []types.ShortURL
		wg   sync.WaitGroup
	)

	if err := json.NewDecoder(req.Body).Decode(&urls); err != nil {
		s.logger.Error("failed to decode request body", "error", err)
		return
	}
	wg.Add(len(urls))
	for _, url := range urls {
		go func(url types.ShortURL) {
			defer wg.Done()
			if err := s.Stor.DeleteUserURL(userID, url); err != nil {
				s.logger.Error("failed to delete URL", "error", err)
			}
		}(url)
	}

	resp.WriteHeader(http.StatusAccepted)
}
