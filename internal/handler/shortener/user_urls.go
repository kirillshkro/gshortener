package shortener

import (
	"encoding/json"
	"net/http"

	"github.com/kirillshkro/gshortener/internal/types"
)

type Getter interface {
	GetUserURLs(resp http.ResponseWriter, req *http.Request)
}

func (s Service) GetUserURLs(resp http.ResponseWriter, req *http.Request) {
	var (
		userID string
		ok     bool
	)
	userID, ok = req.Context().Value(types.UserID).(string)
	if !ok {
		resp.WriteHeader(http.StatusInternalServerError)
		return
	}
	urls, err := s.Stor.GetUserURLs(userID)
	if err != nil {
		resp.WriteHeader(http.StatusNoContent)
		return
	}
	if len(urls) == 0 {
		resp.WriteHeader(http.StatusNoContent)
		s.createCookie(resp)
		return
	}
	var userURLs []types.UserURL
	for _, url := range urls {
		userURLs = append(userURLs, types.UserURL{
			ShortURL:    string(s.ResultAddr) + "/" + url.ShortURL,
			OriginalURL: url.OriginalURL,
		})
	}
	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(http.StatusOK)
	if err = json.NewEncoder(resp).Encode(userURLs); err != nil {
		s.logger.Error("cannot encode response: ", "error: ", err.Error())
		return
	}

}
