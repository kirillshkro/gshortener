package shortener

import (
	"encoding/json"
	"net/http"

	"github.com/kirillshkro/gshortener/internal/handler/shortener/claims"
	"github.com/kirillshkro/gshortener/internal/types"
)

type Getter interface {
	GetUserURLs(resp http.ResponseWriter, req *http.Request)
}

func (s Service) GetUserURLs(resp http.ResponseWriter, req *http.Request) {

	if cookieExist(req, "auth_cookie") {
		cookie, err := req.Cookie("auth_cookie")
		if err != nil {
			resp.WriteHeader(http.StatusNoContent)
			return
		}
		//Получаем токен из cookie
		//и проверяем его на валидность
		token := cookie.Value
		if token == "" {
			resp.WriteHeader(http.StatusUnauthorized)
			return
		}

		userID, err := claims.GetUserID(token)
		if err != nil {
			resp.WriteHeader(http.StatusUnauthorized)
			return
		}
		//Получаем все URL пользователя по его ID
		urls, err := s.Stor.GetUserURLs(userID)
		if err != nil {
			resp.WriteHeader(http.StatusNoContent)
			return
		}
		if len(urls) == 0 {
			s.refreshUserCookie(resp)
			return
		}

		//Отдаем пользователю все URL
		resp.Header().Set("Content-Type", "application/json")
		resp.WriteHeader(http.StatusOK)
		var userURLs []types.UserURL
		for _, url := range urls {
			userURLs = append(userURLs, types.UserURL{
				ShortURL:    string(s.ResultAddr) + "/" + url.ShortURL,
				OriginalURL: url.OriginalURL,
			})
		}
		if err = json.NewEncoder(resp).Encode(userURLs); err != nil {
			s.logger.Error("cannot encode response: ", "error: ", err.Error())
			return
		}
	} else {
		//обновим cookie
		s.refreshUserCookie(resp)
	}
}
