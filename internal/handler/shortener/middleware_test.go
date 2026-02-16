package shortener

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/rs/zerolog"
	"github.com/stretchr/testify/assert"
)

func Test_HandlerWithLog(t *testing.T) {
	setUp()
	defer tearDown()
	testData := RequestData{
		URL: "https://weather.yandex.ru",
	}

	body, err := json.Marshal(testData)
	if err != nil {
		t.Fatal(err)
	}

	testHandler := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
	})

	// Оборачиваем его в middleware
	handler := HandlerWithLog(testHandler)

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader(body))
	w := httptest.NewRecorder()

	// Вызываем обработчик
	handler.ServeHTTP(w, req)

	assert.Equal(t, http.StatusCreated, w.Code)

	assert.Contains(t, "URI request:", w.Body.String())
	assert.Contains(t, "method", w.Body.String())
	assert.Contains(t, "time", w.Body.String())
	assert.Contains(t, "Content length", w.Body.String())
	assert.Contains(t, "status code", w.Body.String())
}

func setUp() {
	// Подавляем вывод логов в stderr при тестах
	zerolog.SetGlobalLevel(zerolog.Disabled)
}

func tearDown() {
	// Возвращаем вывод логов в stderr после тестов
	zerolog.SetGlobalLevel(zerolog.InfoLevel)
}
