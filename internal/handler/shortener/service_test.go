package shortener

import (
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
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
