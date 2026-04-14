package shortener

import "net/http"

type Deleter interface {
	DeleteUserURLs(resp http.ResponseWriter, req *http.Request)
}

func DeleteUserURLs(resp http.ResponseWriter, req *http.Request) {
}
