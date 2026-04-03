package main

import (
	"context"
	"flag"
	"fmt"
	"log/slog"
	"net/http"
	"os"
	"time"

	"github.com/gorilla/mux"
	"github.com/kirillshkro/gshortener/internal/config"
	"github.com/kirillshkro/gshortener/internal/handler/shortener"
	"github.com/kirillshkro/gshortener/internal/handler/shortener/middleware"
	"github.com/kirillshkro/gshortener/internal/repository/storage"
	"github.com/kirillshkro/gshortener/internal/types"
	"github.com/kirillshkro/gshortener/internal/types/model"
	"gorm.io/driver/postgres"
	"gorm.io/gorm"
	"gorm.io/gorm/logger"
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
		db, err := newDB(cfg.DSN)
		if err != nil {
			logger.Warn("Failed to connect to database, switching to the next option", "error", err)
		} else {
			logger.Info("Using database storage")
			stor = storage.NewURLRepository(db)
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
	//Добавляем хандлеры с логгированием
	router.Use(middleware.HandlerWithLog)
	//Добавляем хандлеры с сжатием траффика
	router.Use(middleware.HandlerWithGzip)
	return router
}

func newDB(dsn string) (*gorm.DB, error) {
	dbLogger := logger.NewSlogLogger(
		slog.New(
			slog.NewJSONHandler(os.Stderr, nil),
		),
		logger.Config{
			LogLevel:             logger.Info,
			SlowThreshold:        500 * time.Millisecond,
			ParameterizedQueries: true,
			Colorful:             true,
		},
	)
	conf := &gorm.Config{
		Logger:         dbLogger,
		PrepareStmt:    true,
		TranslateError: true,
	}
	db, err := gorm.Open(postgres.Open(dsn), conf)
	if err != nil {
		return nil, err
	}
	if err = populateTables(db); err != nil {
		return nil, err
	}
	return db, nil
}

func populateTables(db *gorm.DB) error {
	ctx, cancel := context.WithTimeout(context.Background(), 5*time.Second)
	defer cancel()
	if err := db.WithContext(ctx).AutoMigrate(&model.DataURL{}, &model.UserProfile{}); err != nil {
		return err
	}
	return nil
}
