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
	mux := mux.NewRouter()
	//Добавляем хандлеры с логгированием
	mux.Handle("/", shortener.HandlerWithCompress(
		shortener.HandlerWithLog(
			shortener.EncodeHandler(service)))).Methods(http.MethodPost)
	mux.Handle("/{id}", shortener.HandlerWithCompress(shortener.HandlerWithLog(shortener.DecodeHandler(service)))).Methods(http.MethodGet)
	mux.Handle("/api/shorten", shortener.HandlerWithCompress(
		shortener.HandlerWithLog(shortener.CreateShortURLHandler(service)))).Methods(http.MethodPost)
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
