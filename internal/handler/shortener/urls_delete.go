package shortener

import (
	"encoding/json"
	"net/http"

	"github.com/kirillshkro/gshortener/internal/types"
)

type Deleter interface {
	DeleteUserURLs(resp http.ResponseWriter, req *http.Request)
}

func (s Service) DeleteUserURLs(resp http.ResponseWriter, req *http.Request) {
	ctx := req.Context()
	userID := ctx.Value("user_id").(string)
	var shortIDs []types.ShortURL

	if err := json.NewDecoder(req.Body).Decode(&shortIDs); err != nil {
		http.Error(resp, "error decode ", http.StatusInternalServerError)
		return
	}

	for _, shortID := range shortIDs {
		go s.Stor.DeleteUserURL(userID, shortID)
	}
}
