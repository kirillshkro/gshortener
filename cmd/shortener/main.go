package main

import (
	"flag"
	"fmt"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kirillshkro/gshortener/internal/config"
	"github.com/kirillshkro/gshortener/internal/handler/shortener"
	"github.com/kirillshkro/gshortener/internal/handler/shortener/middleware"
	"github.com/kirillshkro/gshortener/internal/types"
)

var cfg *config.Config

func main() {
	parseFlags()
	service := shortener.NewServiceWithAddrWithAddrShortener(types.RawURL(cfg.Address), types.ShortURL(cfg.ShortedURL))
	router := mux.NewRouter()
	router.Handle("/", middleware.EncodeHandler(service)).Methods(http.MethodPost)
	router.Handle("/ping", middleware.PingHandler(service)).Methods(http.MethodGet)
	router.Handle("/{id}", middleware.DecodeHandler(service)).Methods(http.MethodGet)
	router.Handle("/api/shorten", middleware.CreateShortURLHandler(service)).Methods(http.MethodPost)
	//Добавляем хандлеры с логгированием
	router.Use(middleware.HandlerWithLog)
	//Добавляем хандлеры с сжатием траффика
	router.Use(middleware.HandlerWithGzip)
	if err := http.ListenAndServe(cfg.Address, router); err != nil {
		fmt.Printf("error listen server is %s\n", err.Error())
	}
}

func parseFlags() {
	cfg = config.GetConfig()
	flag.StringVar(&cfg.Address, "a", cfg.Address, "Set base host address service")
	flag.StringVar(&cfg.ShortedURL, "b", cfg.ShortedURL, "Set base shorted url")
	flag.StringVar(&cfg.FileDB, "f", cfg.FileDB, "Set path to database")
	flag.StringVar(&cfg.DSN, "d", cfg.DSN, "Set database connection string")
	flag.Parse()
}
