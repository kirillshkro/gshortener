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
	const testData = "https://practicum.yandex.ru"
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
			uri:    testData,
			hash:   hashing([]byte(testData)),
			code:   http.StatusCreated,
		},
		{
			name:   "GET request",
			method: http.MethodGet,
			uri:    testData,
			hash:   hashing([]byte(testData)),
			code:   http.StatusBadRequest,
		},
		{
			name:   "PUT request",
			method: http.MethodPut,
			uri:    testData,
			hash:   hashing([]byte(testData)),
			code:   http.StatusBadRequest,
		},
		{
			name:   "PATCH request",
			method: http.MethodPatch,
			uri:    testData,
			hash:   hashing([]byte(testData)),
			code:   http.StatusBadRequest,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, _ := json.Marshal(test.uri)
			req := httptest.NewRequest(test.method, "/", bytes.NewBuffer(body))
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
		{
			name:   "Bad POST request",
			method: http.MethodPost,
			uri:    testURL,
			status: http.StatusBadRequest,
		},
		{
			name:   "Bad PUT request",
			method: http.MethodPut,
			uri:    testURL,
			status: http.StatusBadRequest,
		},
		{
			name:   "Bad PATCH request",
			method: http.MethodPatch,
			uri:    testURL,
			status: http.StatusBadRequest,
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			req, err := http.NewRequest(test.method, string(hash), nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()

			mux.ServeHTTP(rr, req)

			// проверка статуса
			assert.Equal(t, test.status, rr.Code)

			// проверяем header Location
			var location string
			valLoc := rr.Header().Get("Location")
			if err = json.NewDecoder(strings.NewReader(valLoc)).Decode(&location); err != nil {
				t.Fatal(err)
			}
			assert.Equal(t, test.uri, location)
		})
	}
}
