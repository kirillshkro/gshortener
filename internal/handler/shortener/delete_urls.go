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

	if cookieExist(req, "auth_cookie") {
		authUser := claims.NewAuthUser(auth.NewAuthConfig())
		userToken, err := authUser.Token()
		if err != nil {
			s.logger.Error("failed to get user token", "error", err)
			return
		}

		userID, err := claims.GetUserID(userToken)
		if err != nil {
			s.logger.Error("failed to get user ID", "error", err)
			return
		}

		var urls []types.ShortURL

		if err := json.NewDecoder(req.Body).Decode(&urls); err != nil {
			s.logger.Error("failed to decode request body", "error", err)
			return
		}

		var wg sync.WaitGroup
		wg.Add(len(urls))

		for _, url := range urls {
			defer wg.Done()
			go s.Stor.DeleteUserURL(userID, url)
			s.logger.Info("deleted user URL", "url", url)
		}
		wg.Wait()
	}

	resp.WriteHeader(http.StatusAccepted)
}
