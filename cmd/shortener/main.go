package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kirillshkro/gshortener/internal/config"
	"github.com/kirillshkro/gshortener/internal/handler/shortener"
	"github.com/kirillshkro/gshortener/internal/types"
)

var cfg *config.Config

func main() {
	parseFlags()
	service := shortener.NewServiceWithAddrWithAddrShortener(types.RawURL(cfg.Address), types.ShortURL(cfg.ShortedURL))
	router := mux.NewRouter()
	//Добавляем хандлеры с логгированием
	router.Handle("/", shortener.HandlerWithLog(shortener.EncodeHandler(service))).Methods(http.MethodPost)
	router.Handle("/{id}", shortener.HandlerWithLog(shortener.DecodeHandler(service))).Methods(http.MethodGet)
	router.Handle("/api/shorten", shortener.HandlerWithLog(shortener.CreateShortURLHandler(service))).Methods(http.MethodPost)
	router.Use(shortener.HandlerWithGzip)
	if err := http.ListenAndServe(cfg.Address, router); err != nil {
		fmt.Printf("error listen server is %s\n", err.Error())
	}
}

func parseFlags() {
	cfg = config.NewConfig()
	flag.StringVar(&cfg.Address, "a", cfg.Address, "Set base host address service")
	flag.StringVar(&cfg.ShortedURL, "b", cfg.ShortedURL, "Set base shorted url")
	flag.Parse()
}
