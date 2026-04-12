package shortener

import (
	"encoding/json"
	"net/http"

	"github.com/kirillshkro/gshortener/internal/handler/shortener/claims"
	"github.com/kirillshkro/gshortener/internal/types"
)

type Deleter interface {
	DeleteUserURLs(resp http.ResponseWriter, req *http.Request)
}

func (s Service) DeleteUserURLs(resp http.ResponseWriter, req *http.Request) {
	if cookieExist(req, "auth_cookie") {
		cookie, err := req.Cookie("auth_cookie")
		if err != nil {
			resp.WriteHeader(http.StatusNoContent)
			s.logger.Error("cannot get cookie: ", "error ", err.Error())
			return
		}
		if cookie.Value == "" {
			resp.WriteHeader(http.StatusUnauthorized)
			return
		}
		userID, err := claims.GetUserID(cookie.Value)
		if err != nil {
			resp.WriteHeader(http.StatusUnauthorized)
		}
		var urls []types.ShortURL
		if err := json.NewDecoder(req.Body).Decode(&urls); err != nil {
			resp.WriteHeader(http.StatusBadRequest)
			return
		}
		if err := s.Stor.DeleteUserURLs(userID, urls); err != nil {
			resp.WriteHeader(http.StatusNoContent)
			return
		}
		resp.WriteHeader(http.StatusAccepted)
	} else {
		s.refreshUserCookie(resp)
		return
	}
}
