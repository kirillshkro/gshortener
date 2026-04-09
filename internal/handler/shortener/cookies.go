package shortener

import (
	"io"
	"net/http"
)

func cookieExist(req *http.Request, cookieName string) bool {
	_, err := req.Cookie(cookieName)
	if err != nil {
		return false
	}
	req.Body = io.NopCloser(req.Body)
	return true
}
