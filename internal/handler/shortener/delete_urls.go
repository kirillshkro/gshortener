package shortener

import (
	"encoding/json"
	"net/http"
	"time"

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

	userCookie, err := req.Cookie("auth_cookie")
	if err != nil {
		s.logger.Error("failed to get auth cookie", "error", err)
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}

	userToken := userCookie.Value
	if userToken == "" {
		s.logger.Error("user ID not found")
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
	)

	if err := json.NewDecoder(req.Body).Decode(&urls); err != nil {
		s.logger.Error("failed to decode request body", "error", err)
		return
	}

	errChan := make(chan error, len(urls))

	for _, url := range urls {
		go func(url types.ShortURL) {
			errChan <- s.Stor.DeleteUserURL(userID, url)
		}(url)
	}

	go func() {
		timeout := time.After(10 * time.Second)
		errCount := 0
		for i := 0; i < len(urls); i++ {
			select {
			case err := <-errChan:
				if err != nil {
					s.logger.Error("failed to delete url", "url", string(urls[i]), "error", err)
					errCount++
				}
			case <-timeout:
				s.logger.Error("delete urls timeout")
				return
			}
		}
	}()
	resp.WriteHeader(http.StatusAccepted)
}
