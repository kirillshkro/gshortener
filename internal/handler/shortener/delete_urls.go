package shortener

import (
	"context"
	"encoding/json"
	"net/http"
	"time"

	"github.com/kirillshkro/gshortener/internal/types"
)

type Deleter interface {
	DeleteUserURLs(resp http.ResponseWriter, req *http.Request)
}

func (s Service) DeleteUserURLs(resp http.ResponseWriter, req *http.Request) {
	var (
		userID string
		ok     bool
	)
	userID, ok = req.Context().Value(types.UserID).(string)
	if !ok {
		resp.WriteHeader(http.StatusUnauthorized)
		return
	}

	var (
		urls []types.ShortURL
	)

	ctx := context.WithValue(context.Background(), types.UserID, userID)

	if err := json.NewDecoder(req.Body).Decode(&urls); err != nil {
		s.logger.Error("failed to decode request body", "error", err)
		return
	}

	errChan := make(chan error, len(urls))

	for _, url := range urls {
		go func(url types.ShortURL) {
			errChan <- s.Stor.DeleteUserURL(ctx, url)
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
