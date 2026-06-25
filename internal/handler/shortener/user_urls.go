package shortener

import (
	"encoding/json"
	"net/http"

	"github.com/kirillshkro/gshortener/internal/types"
)

// Getter interface defines the contract for getting user URLs
type Getter interface {
	// GetUserURLs handles the retrieval of user URLs
	GetUserURLs(resp http.ResponseWriter, req *http.Request)
}

// GetUserURLs handles the retrieval of user URLs
// It returns a JSON array of user URLs with their short and original URLs
// Returns HTTP 200 OK on success, HTTP 204 No Content if no URLs found
func (s Service) GetUserURLs(resp http.ResponseWriter, req *http.Request) {
	var (
		userID string
		ok     bool
	)
	userID, ok = req.Context().Value(types.UserID).(string)
	if !ok {
		resp.WriteHeader(http.StatusNoContent)
		return
	}
	urls, err := s.Stor.GetUserURLs(userID)
	if err != nil {
		resp.WriteHeader(http.StatusUnauthorized)
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
