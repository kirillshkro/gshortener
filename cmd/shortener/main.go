package main

import (
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kirillshkro/gshortener/internal/handler/shortener"
)

func main() {
	mux := mux.NewRouter()
	//Добавляем хандлеры
	mux.HandleFunc("/", shortener.URLEncode)
	mux.HandleFunc("/{id}", shortener.URLDecode).Methods("GET")
	if err := http.ListenAndServe(":8080", mux); err != nil {
		fmt.Printf("error listen server is %s\n", err.Error())
	}
}
