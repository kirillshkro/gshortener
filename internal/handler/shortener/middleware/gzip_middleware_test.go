package middleware_test

import (
	"bytes"
	"compress/gzip"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kirillshkro/gshortener/internal/handler/shortener/middleware"
	"github.com/stretchr/testify/suite"
)

type GzipMiddlewareTestSuite struct {
	suite.Suite
}

func (s *GzipMiddlewareTestSuite) SetupTest() {
}

func TestGzipMiddleware(t *testing.T) {
	suite.Run(t, new(GzipMiddlewareTestSuite))
}

// TestGzipMiddleware_SuccessfulCompression проверяет, что middleware успешно сжимает данные.
func (s *GzipMiddlewareTestSuite) TestGzipMiddleware_SuccessfulCompression() {
	// Подготовка данных
	data := []byte(`{"message": "Hello, World!"}`)
	gzData, err := compressToGzip(data)
	s.NoError(err)

	// Создание тестового сервера с middleware
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(gzData)
	})

	handler := middleware.HandlerWithGzip(mux)

	// Подготовка HTTP запроса с заголовком Content-Encoding: gzip
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	s.NoError(err)
	req.Header.Set("Content-Encoding", "gzip")

	// Выполнение запроса
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	// Проверка результата
	resp := recorder.Result()
	s.Equal(http.StatusOK, resp.StatusCode)

	// Распаковка сжатых данных и проверка содержимого
	uncompressedData, err := gzip.NewReader(resp.Body)
	s.NoError(err)
	defer uncompressedData.Close()

	body, err := io.ReadAll(uncompressedData)
	s.NoError(err)
	s.JSONEq(string(data), string(body))
}

// TestGzipMiddleware_UnsuccessfulCompression проверяет, что middleware корректно обрабатывает несжатые данные.
func (s *GzipMiddlewareTestSuite) TestGzipMiddleware_UnsuccessfulCompression() {
	// Подготовка данных
	data := []byte(`{"message": "Hello, World!"}`)

	// Создание тестового сервера с middleware
	mux := http.NewServeMux()
	mux.HandleFunc("/", func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusOK)
		w.Write(data)
	})

	handler := middleware.HandlerWithGzip(mux)

	// Выполнение запроса без заголовка Content-Encoding: gzip
	req, err := http.NewRequest(http.MethodGet, "/", nil)
	s.NoError(err)

	// Выполнение запроса
	recorder := httptest.NewRecorder()
	handler.ServeHTTP(recorder, req)

	// Проверка результата
	resp := recorder.Result()
	s.Equal(http.StatusOK, resp.StatusCode)

	// Проверка содержимого ответа
	body, err := io.ReadAll(resp.Body)
	s.NoError(err)
	s.JSONEq(string(data), string(body))
}

func compressToGzip(data []byte) ([]byte, error) {
	var buf bytes.Buffer
	gzw := gzip.NewWriter(&buf)
	if _, err := gzw.Write(data); err != nil {
		return nil, err
	}
	if err := gzw.Close(); err != nil {
		return nil, err
	}
	return buf.Bytes(), nil
}
