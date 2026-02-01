package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kirillshkro/gshortener/internal/config"
	"github.com/kirillshkro/gshortener/internal/handler/shortener"
)

var cfg *config.Config

func main() {
	service := shortener.NewService()
	parseFlags()
	mux := mux.NewRouter()
	//Добавляем хандлеры
	mux.HandleFunc("/", service.URLEncode)
	mux.HandleFunc("/{id}", service.URLDecode).Methods(http.MethodGet)
	if err := http.ListenAndServe(cfg.Address, mux); err != nil {
		fmt.Printf("error listen server is %s\n", err.Error())
	}
}

func parseFlags() {
	cfg = config.NewConfig()
	flag.StringVar(&cfg.Address, "a", cfg.Address, "Set base host address service")
	flag.StringVar(&cfg.ShortedURL, "b", cfg.ShortedURL, "Set base shorted url")
	flag.Parse()
}
