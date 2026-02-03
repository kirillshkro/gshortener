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
	"github.com/kirillshkro/gshortener/internal/repository/storage"
	"github.com/stretchr/testify/assert"
)

var (
	service  *Service
	router   *mux.Router      = mux.NewRouter()
	fakeServ *httptest.Server = httptest.NewServer(router)
)

func setup() {
	service = NewServiceWithAddrWithAddrShortener(storage.RawURL(fakeServ.URL), storage.ShortURL(fakeServ.URL))
	router.HandleFunc("/", service.URLEncode).Methods(http.MethodPost)
	router.HandleFunc("/{id}", service.URLDecode).Methods(http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch)
}

func Test_URLEncode(t *testing.T) {
	setup()
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
			hash:   Hashing([]byte(testData)),
			code:   http.StatusCreated,
		},
		{
			name:   "GET request",
			method: http.MethodGet,
			uri:    testData,
			hash:   Hashing([]byte(testData)),
			code:   http.StatusBadRequest,
		},
		{
			name:   "PUT request",
			method: http.MethodPut,
			uri:    testData,
			hash:   Hashing([]byte(testData)),
			code:   http.StatusBadRequest,
		},
		{
			name:   "PATCH request",
			method: http.MethodPatch,
			uri:    testData,
			hash:   Hashing([]byte(testData)),
			code:   http.StatusBadRequest,
		},
	}
	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			body, _ := json.Marshal(test.uri)
			req := httptest.NewRequest(test.method, "/", bytes.NewBuffer(body))
			recorder := httptest.NewRecorder()
			service.URLEncode(recorder, req)
			if recorder.Code != test.code {
				t.Errorf("test failed because expected code: %d, real code: %d\n", test.code, recorder.Code)
			}
		})
	}
}

func Test_URLDecode(t *testing.T) {
	setup()

	const testURL = `https://practicum.yandex.ru`
	body, err := json.Marshal(testURL)
	if err != nil {
		t.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, fakeServ.URL, strings.NewReader(testURL))
	rr := httptest.NewRecorder()
	service.URLEncode(rr, req)
	resp := rr.Result()
	defer resp.Body.Close()
	body, _ = io.ReadAll(resp.Body)
	hashed, _ := strings.CutPrefix(string(body), fakeServ.URL+"/")
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
			req := httptest.NewRequest(test.method, fakeServ.URL+"/"+hashed, nil)
			rr = httptest.NewRecorder()
			service.URLDecode(rr, req)
			resp = rr.Result()
			defer resp.Body.Close()
			// проверка статуса
			assert.Equal(t, test.status, resp.StatusCode)
			if req.Method == http.MethodGet {
				location := resp.Header.Get("Location")
				t.Log("Location: ", location)
				assert.Equal(t, testURL, location)
			}
		})
	}
	fakeServ.Close()
}
