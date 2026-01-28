package shortener

import (
	"crypto/sha1"
	"encoding/hex"
	"log"
	"net/http"
)

type URLString = string

var urls map[URLString]string = make(map[URLString]string)

// Принимает на вход URL, возвращает базовый URL сервиса + хэш исходного URL
func URLEncode(resp http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}
	baseURL := "http://" + req.Host
	resp.Header().Set("Content-Type", "text/plain")
	resp.WriteHeader(http.StatusCreated)

	//получить URI
	content := req.URL.Path
	hasher := sha1.New()
	//взять от него хэш
	raw := hasher.Sum([]byte(content))
	shorted := hex.EncodeToString(raw[:])
	outURL := req.URL.Scheme + string(shorted[:8])

	body := baseURL + req.URL.Path
	_, err := resp.Write([]byte(body))
	if err != nil {
		log.Printf("don't send response because by %s\n", err.Error())
	}
	if _, ok := urls[outURL]; !ok {
		urls[outURL] = body
	}
	log.Println("body = ", body)
}

func URLDecode(resp http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}
	pattern := req.RequestURI[1:]
	original := urls[pattern]
	resp.Header().Set("Location", original)
	resp.WriteHeader(http.StatusTemporaryRedirect)
}
