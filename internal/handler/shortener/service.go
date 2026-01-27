package shortener

import (
	"log"
	"net/http"
)

func URLEncode(resp http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}
	resp.Header().Set("Content-Type", "text/plain")
	//resp.Header().Add("Content-Length", "30")
	resp.WriteHeader(http.StatusCreated)

	body := []byte("http://localhost:8080/EwHXdJfB")
	_, err := resp.Write(body)
	if err != nil {
		log.Printf("don't send response because by %s\n", err.Error())
	}
}

func URLDecode(resp http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}
	resp.Header().Set("Location", "https://practicum.yandex.ru")
	resp.WriteHeader(http.StatusTemporaryRedirect)
}
