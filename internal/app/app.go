package app

import (
	"context"
	"errors"
	"flag"
	"log"
	"log/slog"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/kirillshkro/gshortener/internal/config"
	"github.com/kirillshkro/gshortener/internal/handler/shortener"
	"github.com/kirillshkro/gshortener/internal/handler/shortener/middleware"
	"github.com/kirillshkro/gshortener/internal/repository/storage"
	"github.com/kirillshkro/gshortener/internal/service/audit"
	"github.com/kirillshkro/gshortener/internal/types"
)

type App struct {
	cfg       *config.Config
	service   *shortener.Service
	router    *mux.Router
	server    *http.Server
	interrupt chan os.Signal
}

func NewApp(cfg *config.Config) (*App, error) {
	app := &App{
		cfg:       cfg,
		interrupt: make(chan os.Signal, 1),
	}
	return app, nil
}

func (a *App) setupService() (*shortener.Service, error) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	var (
		stor    storage.IStorage
		err     error
		subject audit.Subject
	)
	service := shortener.NewServiceWithAddrWithAddrShortener(types.RawURL(a.cfg.Address), types.ShortURL(a.cfg.ShortedURL))

	if a.cfg.DSN == "" && a.cfg.FileDB == "" {
		service.Stor = storage.NewMemoryStorage()
		logger.Info("All external storages are unavailable. Using in-memory storage (will not be saved after restart)")
	}

	if a.cfg.FileDB != "" {
		if stor, err = storage.GetFileStorage(a.cfg.FileDB); err != nil {
			logger.Warn("Failed to use file storage, switching to the next option", "error", err)
		} else {
			if service.Stor == nil {
				service.Stor = stor
			}
			logger.Info("Using file storage")
		}
	}

	if a.cfg.DSN != "" {
		logger.Info("Try connect to database", "dsn", a.cfg.DSN)
		if stor, err = storage.GetDBStorage(a.cfg.DSN); err != nil {
			logger.Warn("Failed to connect to database, switching to the next option", "error", err)
		} else {
			logger.Info("Using database storage")
			service.Stor = stor
		}
	}

	if a.cfg.AuditFile != "" {
		fileAudit, err := audit.GetAuditService(a.cfg.AuditFile)
		if err != nil {
			logger.Warn("Failed to create audit service", "Warn", err)
		}
		subject.Register(fileAudit)
		service.SetSubject(&subject)
	}

	if a.cfg.AuditURL != "" {
		urlAudit, err := audit.GetNetAuditService(a.cfg.AuditURL)
		if err != nil {
			logger.Warn("Failed to create audit service", "Warn", err)
		}
		subject.Register(urlAudit)
		service.SetSubject(&subject)
	}

	return service, nil
}

func (a *App) setupRouter(service *shortener.Service) *mux.Router {
	router := mux.NewRouter()
	router.Handle("/", middleware.EncodeHandler(service)).Methods(http.MethodPost)
	router.Handle("/ping", middleware.PingHandler(service)).Methods(http.MethodGet)
	router.Handle("/{id}", middleware.DecodeHandler(service)).Methods(http.MethodGet)

	//Добавляем хандлеры с созданием коротких ссылок
	router.Handle("/api/shorten/batch", middleware.BatchCreateURLHandler(service)).Methods(http.MethodPost)
	router.Handle("/api/shorten", middleware.CreateShortURLHandler(service)).Methods(http.MethodPost)

	//Добавляем хандлеры с получением информации о короткой ссылке
	router.Handle("/api/user/urls", middleware.GetUserURLsHandler(service)).Methods(http.MethodGet)
	router.Handle("/api/user/urls", middleware.DeleteUserURLsHandler(service)).Methods(http.MethodDelete)

	//Добавляем middleware с логгированием
	router.Use(middleware.HandlerWithLog)
	router.Use(service.AuthMiddleware)

	//Добавляем middleware с сжатием траффика
	router.Use(middleware.HandlerWithGzip)

	return router
}

func (a *App) parseFlags() {
	flag.StringVar(&a.cfg.Address, "a", a.cfg.Address, "Set base host address service")
	flag.StringVar(&a.cfg.ShortedURL, "b", a.cfg.ShortedURL, "Set base shorted url")
	flag.StringVar(&a.cfg.FileDB, "f", a.cfg.FileDB, "Set path to database")
	flag.StringVar(&a.cfg.DSN, "d", a.cfg.DSN, "Set database connection string")
	flag.StringVar(&a.cfg.AuditURL, "audit-url", a.cfg.AuditURL, "Set audit service url")
	flag.StringVar(&a.cfg.AuditFile, "audit-file", a.cfg.AuditFile, "Set audit file path")
	flag.Parse()

	var err error
	a.service, err = a.setupService()
	if err != nil {
		return
	}

	a.router = a.setupRouter(a.service)
}

func (a *App) runServer() error {
	server := &http.Server{
		Addr:    a.cfg.Address,
		Handler: a.router,
	}
	go func() {
		log.Printf("server is listening on %s\n", a.cfg.Address)
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("error listen server is %s\n", err.Error())
		}
	}()
	a.server = server
	return nil
}

func (a *App) gracefulShutdown() {
	signal.Notify(a.interrupt, syscall.SIGTERM, syscall.SIGINT)
	<-a.interrupt
	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()
	if err := a.server.Shutdown(ctx); err != nil {
		log.Fatalf("Failed to shutdown server: %v", err)
		log.Fatalf("Server stopped")
	}
	log.Println("Shutting down server...")
}

func (a *App) Run() error {
	a.parseFlags()
	err := a.runServer()
	if err != nil {
		return err
	}

	a.gracefulShutdown()
	return nil
}
