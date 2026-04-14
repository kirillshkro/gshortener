package shortener

import "net/http"

type Deleter interface {
	DeleteUserURLs(resp http.ResponseWriter, req *http.Request)
}

func (s Service) DeleteUserURLs(resp http.ResponseWriter, req *http.Request) {
	resp.WriteHeader(http.StatusAccepted)
}
