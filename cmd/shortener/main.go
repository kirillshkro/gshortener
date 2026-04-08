package main

import (
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"

	"github.com/gorilla/mux"
	"github.com/kirillshkro/gshortener/internal/config"
	"github.com/kirillshkro/gshortener/internal/handler/shortener"
	"github.com/kirillshkro/gshortener/internal/handler/shortener/middleware"
	"github.com/kirillshkro/gshortener/internal/repository/storage"
	"github.com/kirillshkro/gshortener/internal/types"
)

var cfg *config.Config

func main() {
	parseFlags()
	service := setupService(cfg)
	router := setupRouter(service)
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
	setupService(cfg)
}

func setupService(cfg *config.Config) *shortener.Service {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	var (
		stor storage.IStorage
		err  error
	)
	service := shortener.NewServiceWithAddrWithAddrShortener(types.RawURL(cfg.Address), types.ShortURL(cfg.ShortedURL))
	if cfg.DSN == "" && cfg.FileDB == "" {
		service.Stor = storage.NewMemoryStorage()
		logger.Info("All external storages are unavailable. Using in-memory storage (will not be saved after restart)")
	}
	if (cfg.FileDB != "") && (cfg.DSN == "") {
		if stor, err = storage.GetFileStorage(cfg.FileDB); err != nil {
			logger.Warn("Failed to use file storage, switching to the next option", "error", err)
		} else {
			if service.Stor == nil {
				service.Stor = stor
			}
			logger.Info("Using file storage")
		}
	}
	if cfg.DSN != "" {
		logger.Info("Try connect to database", "dsn", cfg.DSN)
		if stor, err = storage.GetDBStorage(cfg.DSN); err != nil {
			logger.Warn("Failed to connect to database, switching to the next option", "error", err)
		} else {
			logger.Info("Using database storage")
			service.Stor = stor
		}
	}

	return service
}

func setupRouter(service *shortener.Service) *mux.Router {
	router := mux.NewRouter()
	router.Handle("/", middleware.EncodeHandler(service)).Methods(http.MethodPost)
	router.Handle("/ping", middleware.PingHandler(service)).Methods(http.MethodGet)
	router.Handle("/{id}", middleware.DecodeHandler(service)).Methods(http.MethodGet)
	//Добавляем хандлеры с созданием коротких ссылок
	router.Handle("/api/shorten/batch", middleware.BatchCreateURLHandler(service)).Methods(http.MethodPost)
	router.Handle("/api/shorten", middleware.CreateShortURLHandler(service)).Methods(http.MethodPost)
	//Добавляем хандлеры с получением информации о короткой ссылке
	router.Handle("/api/user/urls", middleware.GetUserURLsHandler(service)).Methods(http.MethodGet)
	//Добавляем middleware с логгированием
	router.Use(middleware.HandlerWithLog)
	//Добавляем middleware с сжатием траффика
	router.Use(middleware.HandlerWithGzip)
	//Добавляем middleware аутентификации
	router.Use(middleware.AuthMiddleware)
	return router
}
