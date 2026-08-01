// Package app предоставляет основную функциональность приложения для сокращения URL.
// Он объединяет конфигурацию, сервисы, маршрутизацию и управление жизненным циклом HTTP-сервера.
package app

import (
	"context"
	"errors"
	"flag"
	"log"
	"log/slog"
	"net"
	"net/http"
	"os"
	"os/signal"
	"sync"
	"syscall"
	"time"

	"github.com/gorilla/mux"
	"github.com/kirillshkro/gshortener/internal/config"
	grpcserver "github.com/kirillshkro/gshortener/internal/grpc"
	"github.com/kirillshkro/gshortener/internal/handler/shortener"
	"github.com/kirillshkro/gshortener/internal/handler/shortener/middleware"
	"github.com/kirillshkro/gshortener/internal/proto"
	"github.com/kirillshkro/gshortener/internal/repository/storage"
	"github.com/kirillshkro/gshortener/internal/service/audit"
	"github.com/kirillshkro/gshortener/internal/types"
	"google.golang.org/grpc"
)

// App представляет основную структуру приложения, объединяющую все компоненты.
// Содержит конфигурацию, сервис сокращения URL, маршрутизатор HTTP,
// HTTP-сервер и канал для обработки сигналов прерывания.
type App struct {
	cfg         *config.Config     // Конфигурация приложения
	service     *shortener.Service // Сервис сокращения URL
	router      *mux.Router        // Маршрутизатор HTTP-запросов
	server      *http.Server       // HTTP-сервер
	grpcServer  *grpc.Server       // gRPC-сервер
	interrupt   chan os.Signal     // Канал для получения сигналов прерывания
	mu          sync.Mutex         // Синхронизатор для безопасного обновления конфигурации
}

// NewApp создает новый экземпляр приложения с переданной конфигурацией.
// Инициализирует канал для сигналов прерывания и сохраняет конфигурацию в файл.
//
// Параметры:
//   - cfg: конфигурация приложения
//
// Возвращает:
//   - *App: указатель на созданное приложение
//   - error: ошибка, если не удалось создать приложение
func NewApp(cfg *config.Config) (*App, error) {
	app := &App{
		cfg:       cfg,
		interrupt: make(chan os.Signal, 1),
		mu:        sync.Mutex{},
	}
	return app, nil
}

// setupService инициализирует сервис сокращения URL с соответствующим хранилищем
// и системами аудита на основе конфигурации приложения.
//
// Процесс настройки:
// 1. Создание JSON-логгера
// 2. Создание сервиса с базовыми URL-адресами
// 3. Последовательный выбор хранилища: память -> файл -> БД
// 4. Настройка системы аудита (файловый и сетевой)
//
// Возвращает:
//   - *shortener.Service: настроенный сервис сокращения URL
//   - error: ошибка, если не удалось создать сервис
func (a *App) setupService() (*shortener.Service, error) {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	var (
		stor    storage.IStorage
		err     error
		subject audit.Subject
	)

	// Создание базового сервиса с начальными URL-адресами
	service := shortener.NewServiceWithAddrWithAddrShortener(types.RawURL(a.cfg.Address), types.ShortURL(a.cfg.ShortedURL))

	// Проверка доступности внешних хранилищ, использование in-memory как запасной вариант
	if a.cfg.DSN == "" && a.cfg.FileDB == "" {
		service.Stor = storage.NewMemoryStorage()
		logger.Info("All external storages are unavailable. Using in-memory storage (will not be saved after restart)")
	}

	// Настройка файлового хранилища
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

	// Настройка хранилища базы данных (имеет приоритет над файловым)
	if a.cfg.DSN != "" {
		logger.Info("Try connect to database", "dsn", a.cfg.DSN)
		if stor, err = storage.GetDBStorage(a.cfg.DSN); err != nil {
			logger.Warn("Failed to connect to database, switching to the next option", "error", err)
		} else {
			logger.Info("Using database storage")
			service.Stor = stor
		}
	}

	// Load config from JSON file if specified via flag
	if a.cfg.ConfigFile != "" {
		if err := a.cfg.ReadFromJSON(a.cfg.ConfigFile); err != nil {
			log.Fatalf("Failed to read config file %q: %v", a.cfg.ConfigFile, err)
		}
	}

	// Настройка системы аудита (наблюдатели)
	service.SetSubject(&subject)

	// Добавление файлового аудита
	if a.cfg.AuditFile != "" {
		fileAudit, err := audit.GetAuditService(a.cfg.AuditFile)
		if err != nil {
			logger.Warn("Failed to create audit service", "Warn", err)
		}
		subject.Register(fileAudit)
	}

	// Добавление сетевого аудита
	if a.cfg.AuditURL != "" {
		urlAudit, err := audit.GetNetAuditService(a.cfg.AuditURL)
		if err != nil {
			logger.Warn("Failed to create audit service", "Warn", err)
		}
		subject.Register(urlAudit)
	}

	return service, nil
}

// setupRouter настраивает маршрутизатор HTTP с обработчиками для всех эндпоинтов.
//
// Регистрирует следующие эндпоинты:
//   - POST / - сокращение URL
//   - GET /ping - проверка доступности сервиса
//   - GET /{id} - получение оригинального URL по короткому идентификатору
//   - POST /api/shorten - создание короткой ссылки (JSON)
//   - POST /api/shorten/batch - пакетное создание коротких ссылок
//   - GET /api/user/urls - получение всех ссылок пользователя
//   - DELETE /api/user/urls - удаление ссылок пользователя
//
// Добавляет middleware:
//   - Логирование запросов
//   - Аутентификация пользователей
//   - Сжатие трафика
//
// Параметры:
//   - service: сервис сокращения URL
//
// Возвращает:
//   - *mux.Router: настроенный маршрутизатор
func (a *App) setupRouter(service *shortener.Service) *mux.Router {
	router := mux.NewRouter()

	// Базовые эндпоинты
	router.HandleFunc("/", service.URLEncode).Methods(http.MethodPost)
	router.HandleFunc("/ping", service.Ping).Methods(http.MethodGet)
	router.HandleFunc("/{id}", service.URLDecode).Methods(http.MethodGet)

	// Эндпоинты для создания коротких ссылок
	router.HandleFunc("/api/shorten/batch", service.BatchCreateShortURL).Methods(http.MethodPost)
	router.HandleFunc("/api/shorten", service.CreateShortURL).Methods(http.MethodPost)

	// Эндпоинты для управления ссылками пользователя
	router.HandleFunc("/api/user/urls", service.GetUserURLs).Methods(http.MethodGet)
	router.HandleFunc("/api/user/urls", service.DeleteUserURLs).Methods(http.MethodDelete)

	// Middleware для всех запросов
	router.Use(middleware.HandlerWithLog)  // Логирование
	router.Use(service.AuthMiddleware)     // Аутентификация
	router.Use(middleware.HandlerWithGzip) // Сжатие

	return router
}

// parseFlags разбирает аргументы командной строки и настраивает приложение.
//
// Поддерживаемые флаги:
//   - -a: адрес сервиса (по умолчанию из конфигурации)
//   - -b: базовый URL для коротких ссылок (по умолчанию из конфигурации)
//   - -f: путь к файлу БД (по умолчанию из конфигурации)
//   - -d: строка подключения к БД (по умолчанию из конфигурации)
//   - -audit-url: URL сервиса аудита (по умолчанию из конфигурации)
//   - -audit-file: путь к файлу аудита (по умолчанию из конфигурации)
//   - -c: путь к файлу конфигурации JSON
//
// После разбора флагов инициализирует сервис и маршрутизатор.
func (a *App) parseFlags() {
	flag.StringVar(&a.cfg.Address, "a", a.cfg.Address, "Set base host address service")
	flag.StringVar(&a.cfg.ShortedURL, "b", a.cfg.ShortedURL, "Set base shorted url")
	flag.StringVar(&a.cfg.FileDB, "f", a.cfg.FileDB, "Set path to database")
	flag.StringVar(&a.cfg.DSN, "d", a.cfg.DSN, "Set database connection string")
	flag.StringVar(&a.cfg.AuditURL, "audit-url", a.cfg.AuditURL, "Set audit service url")
	flag.StringVar(&a.cfg.AuditFile, "audit-file", a.cfg.AuditFile, "Set audit file path")
	flag.BoolVar(&a.cfg.EnableHTTPS, "s", a.cfg.EnableHTTPS, "Set HTTPS mode")
	flag.StringVar(&a.cfg.ConfigFile, "c", a.cfg.ConfigFile, "Set path to JSON configuration file")
	flag.StringVar(&a.cfg.GRPCAddress, "grpc", a.cfg.GRPCAddress, "Set gRPC server address")
	flag.Parse()

	var err error
	a.service, err = a.setupService()
	if err != nil {
		return
	}

	a.router = a.setupRouter(a.service)
}

// setupGRPCServer настраивает gRPC-сервер с зарегистрированным ShortenerService.
// Возвращает настроенный grpc.Server или nil, если GRPC_ADDRESS не задан в конфигурации.
func (a *App) setupGRPCServer() *grpc.Server {
	if a.cfg.GRPCAddress == "" {
		return nil
	}

	s := grpc.NewServer()
	proto.RegisterShortenerServiceServer(s, grpcserver.NewServer(a.service))

	return s
}

// runGRPCServer запускает gRPC-сервер в отдельной горутине.
// При возникновении ошибки выполнения логирует фатальную ошибку.
func (a *App) runGRPCServer() {
	if a.grpcServer == nil {
		return
	}

	lis, err := net.Listen("tcp", a.cfg.GRPCAddress)
	if err != nil {
		log.Fatalf("failed to listen for gRPC on %s: %v", a.cfg.GRPCAddress, err)
	}

	go func() {
		log.Printf("gRPC server is listening on %s\n", a.cfg.GRPCAddress)
		if err := a.grpcServer.Serve(lis); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			log.Fatalf("gRPC server error: %v", err)
		}
	}()
}

// runServer запускает HTTP-сервер в отдельной горутине.
// При возникновении ошибки выполнения сервера (кроме планового завершения)
// логирует фатальную ошибку.
//
// Возвращает:
//   - error: ошибка, если не удалось запустить сервер
func (a *App) runServer() error {
	server := &http.Server{
		Addr:    a.cfg.Address,
		Handler: a.router,
	}

	go func() {
		log.Printf("HTTP server is listening on %s\n", a.cfg.Address)
		if !a.cfg.EnableHTTPS {
			if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatalf("error listen server is %s\n", err.Error())
			}
		} else {
			if err := server.ListenAndServeTLS(a.cfg.CertFile, a.cfg.KeyFile); err != nil && !errors.Is(err, http.ErrServerClosed) {
				log.Fatalf("error listen server is %s\n", err.Error())
			}
		}
	}()

	a.server = server
	return nil
}

// gracefulShutdown ожидает сигналы завершения и выполняет корректное завершение серверов.
//
// Ожидает сигналы SIGTERM, SIGQUIT и SIGINT.
// При получении сигнала создает контекст с таймаутом в 30 секунд
// для корректного завершения HTTP и gRPC соединений.
// При ошибке завершения логирует фатальную ошибку.
func (a *App) gracefulShutdown() {
	signal.Notify(a.interrupt, syscall.SIGTERM, syscall.SIGINT, syscall.SIGQUIT)
	<-a.interrupt

	ctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if a.server != nil {
		if err := a.server.Shutdown(ctx); err != nil {
			log.Printf("Failed to shutdown HTTP server: %v", err)
		}
	}

	if a.grpcServer != nil {
		a.grpcServer.GracefulStop()
	}

	log.Println("Shutting down servers...")
}

// Run запускает основную логику приложения:
// 1. Разбор флагов командной строки
// 2. Настройка и запуск HTTP-сервера
// 3. Настройка и запуск gRPC-сервера
// 4. Ожидание сигналов завершения
// 5. Корректное завершение работы
//
// Возвращает:
//   - error: ошибка, если не удалось запустить сервер
func (a *App) Run() error {
	a.parseFlags()
	err := a.runServer()
	if err != nil {
		return err
	}

	a.grpcServer = a.setupGRPCServer()
	a.runGRPCServer()

	a.gracefulShutdown()
	return nil
}
