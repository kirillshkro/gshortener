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
	if req.Method != http.MethodGet {
		resp.WriteHeader(http.StatusMethodNotAllowed)
		return
	}
	cookie, err := req.Cookie("auth_cookie")
	if err != nil {
		resp.WriteHeader(http.StatusUnauthorized)
		return
	}
	if cookie.Value == "" {
		resp.WriteHeader(http.StatusUnauthorized)
		return
	}
	//Получаем UserID из cookie
	userUUID := cookie.Value
	//Получаем все URL пользователя по его ID
	urls, err := s.Stor.GetUserURLs(userUUID)
	if err != nil {
		resp.WriteHeader(http.StatusNoContent)
		return
	}
	//Отдаем пользователю все URL
	resp.WriteHeader(http.StatusOK)
	var userURLs []types.UserURL
	for _, url := range urls {
		userURLs = append(userURLs, types.UserURL{
			ShortURL:    url.ShortURL,
			OriginalURL: url.OriginalURL,
		})
	}
	resp.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(resp).Encode(userURLs); err != nil {
		s.logger.Error("cannot encode response: ", "error: ", err.Error())
		return
	}
}
