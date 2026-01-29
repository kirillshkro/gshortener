package shortener

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/gorilla/mux"
	"github.com/stretchr/testify/assert"
)

func Test_URLEncode(t *testing.T) {
	tests := []struct {
		name   string
		method string
		uri    string
		hash   string
		code   int
	}{
		{
			name:   "POST request",
			method: http.MethodPost,
			uri:    "https://practicum.yandex.ru",
			hash:   hashing([]byte("https://practicum.yandex.ru")),
			code:   http.StatusCreated,
		},
		{
			name:   "GET request",
			method: http.MethodGet,
			uri:    "https://practicum.yandex.ru",
			hash:   hashing([]byte("https://practicum.yandex.ru")),
			code:   http.StatusBadRequest,
		},
		{
			name:   "PUT request",
			method: http.MethodPut,
			uri:    "https://practicum.yandex.ru",
			hash:   hashing([]byte("https://practicum.yandex.ru")),
			code:   http.StatusBadRequest,
		},
		{
			name:   "PATCH request",
			method: http.MethodPatch,
			uri:    "https://practicum.yandex.ru",
			hash:   hashing([]byte("https://practicum.yandex.ru")),
			code:   http.StatusBadRequest,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body := strings.NewReader(`{` + test.uri + `}`)
			req := httptest.NewRequest(test.method, "/", body)
			recorder := httptest.NewRecorder()
			URLEncode(recorder, req)
			if recorder.Code != test.code {
				t.Errorf("test failed because expected code: %d, real code: %d\n", test.code, recorder.Code)
			}
		})
	}
}

func Test_URLDecode(t *testing.T) {
	const baseURL = "http://localhost:8080/"
	const testURL = "https://practicum.yandex.ru"
	mux := mux.NewRouter()
	mux.HandleFunc("/", URLEncode).Methods(http.MethodPost)
	mux.HandleFunc("/{id}", URLDecode).Methods(http.MethodGet)
	server := httptest.NewServer(mux)
	defer server.Close()
	body, _ := json.Marshal(testURL)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body))
	defer req.Body.Close()

	resp, err := http.Post(server.URL, "application/json", req.Body)
	if err != nil {
		t.Error("Test failed")
	}
	defer resp.Body.Close()
	hash, err := io.ReadAll(resp.Body)
	if err != nil {
		t.Errorf("test failed because error: %s\n", err.Error())
	}

	tests := []struct {
		name   string
		method string
		uri    string
		status int
	}{
		{
			name:   "Normal GET request",
			method: http.MethodGet,
			uri:    testURL,
			status: http.StatusTemporaryRedirect,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, err := http.NewRequest(http.MethodGet, string(hash), nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()

			mux.ServeHTTP(rr, req)

			// проверка статуса
			if status := rr.Code; status != http.StatusTemporaryRedirect {
				t.Errorf("expected status %d, got %d", http.StatusTemporaryRedirect, status)
			}

			// проверяем header Location
			location := rr.Header().Get("Location")
			assert.Equal(t, test.uri, location)
		})
	}
}
