package shortener

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
)

type URLString = string

var urls map[URLString]string = make(map[URLString]string)

// Принимает на вход URL, возвращает базовый URL сервиса + хэш исходного URL
func URLEncode(resp http.ResponseWriter, req *http.Request) {
	const baseURL = "http://localhost:8080/"
	if req.Method != http.MethodPost {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}
	defer req.Body.Close()
	bodyReq, err := io.ReadAll(req.Body)
	if err != nil {
		log.Println("cannot read request: ", err.Error())
	}
	resp.Header().Set("Content-Type", "text/plain")
	resp.WriteHeader(http.StatusCreated)
	content := hashing(bodyReq)
	outData := baseURL + content
	if _, ok := urls[content]; !ok {
		urls[content] = string(bodyReq)
	}
	if _, err = resp.Write([]byte(outData)); err != nil {
		log.Printf("don't send response because by %s\n", err.Error())
	}
}

func URLDecode(resp http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}
	vars := mux.Vars(req)
	id := vars["id"]
	pattern := id
	original := urls[pattern]
	resp.Header().Set("Location", original)
	resp.WriteHeader(http.StatusTemporaryRedirect)
}

func hashing(data []byte) string {
	hashed := sha1.Sum(data)
	shorthed := hashed[:6]
	content := hex.EncodeToString(shorthed)
	return content
}
