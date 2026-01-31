package shortener

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

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
	const testURL = `https://practicum.yandex.ru`
	body, _ := json.Marshal(testURL)
	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewBuffer(body))
	rr := httptest.NewRecorder()
	URLEncode(rr, req)
	resp := rr.Result()
	body, _ = io.ReadAll(resp.Body)
	hashed, _ := strings.CutPrefix(string(body), baseURL)
	resp.Body.Close()
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
			rr = httptest.NewRecorder()
			req, err := http.NewRequest(test.method, "/"+hashed, nil)
			if err != nil {
				t.Fatal(err)
			}
			URLDecode(rr, req)
			resp := rr.Result()
			defer resp.Body.Close()

			// проверка статуса
			assert.Equal(t, test.status, resp.StatusCode)
			if req.Method == http.MethodGet {

				location := resp.Header.Get("Location")
				t.Log("Location: ", location)

			}
		})
	}
}
