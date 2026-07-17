package shortener

import (
	"database/sql"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kirillshkro/gshortener/internal/config"
	"github.com/kirillshkro/gshortener/internal/types"
	"github.com/stretchr/testify/suite"
)

type PingTestSuite struct {
	suite.Suite
	service *Service
	db      *sql.DB
}

func (s *PingTestSuite) SetupTest() {
	var err error
	// Настройка конфигурации с базой данных
	cfg := config.GetConfig()
	cfg.DSN = "postgres://postgres@localhost:5432?sslmode=disable"

	// Создание тестовой базы данных
	s.db, err = sql.Open("postgres", cfg.DSN)
	if err != nil {
		s.T().Fatal(err)
	}
	defer s.db.Close()

	// Удаление таблиц перед каждым тестом
	s.db.Exec("DROP TABLE IF EXISTS urls;")

	// Создание таблицы для тестов
	s.db.Exec(`
		CREATE TABLE urls (
			short_url TEXT PRIMARY KEY,
			original_url TEXT NOT NULL,
			user_uuid TEXT
		);
	`)

	// Инициализация сервиса с базой данных
	s.service = NewServiceWithAddr(types.RawURL(cfg.Address))
}

func (s *PingTestSuite) TearDownTest() {
	// Удаление тестовой базы данных после каждого теста
	s.db.Exec("DROP TABLE IF EXISTS urls;")
}

func TestPingTestSuite(t *testing.T) {
	suite.Run(t, new(PingTestSuite))
}

func (s *PingTestSuite) TestPing_HealthyDBConnection() {
	// Создаем тестируемый HTTP-запрос
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)

	// Создаем ответный объект
	w := httptest.NewRecorder()

	// Выполняем метод Ping сервиса
	s.service.Ping(w, req)

	// Проверяем статус код ответа
	s.Require().Equal(http.StatusOK, w.Code)
}

func (s *PingTestSuite) TestPing_InvalidMethod() {
	// Создаем тестируемый HTTP-запрос с некорректным методом
	req := httptest.NewRequest(http.MethodPost, "/ping", nil)

	// Создаем ответный объект
	w := httptest.NewRecorder()

	// Выполняем метод Ping сервиса
	s.service.Ping(w, req)

	// Проверяем статус код ответа
	s.Require().Equal(http.StatusMethodNotAllowed, w.Code)
}
