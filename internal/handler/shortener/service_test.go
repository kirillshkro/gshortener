package shortener

import (
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
	router.HandleFunc("/", service.URLEncode).Methods(http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch)
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
			req, err := http.NewRequest(test.method, fakeServ.URL+"/", strings.NewReader(string(body)))
			if err != nil {
				t.Fatal(err)
			}
			client := fakeServ.Client()
			resp, err := client.Do(req)
			if err != nil {
				t.Fatal(err)
			}
			defer resp.Body.Close()
			if resp.StatusCode != test.code {
				t.Errorf("test failed because expected code: %d, real code: %d\n", test.code, resp.StatusCode)
			}
		})
	}
}

func Test_URLDecode(t *testing.T) {
	setup()
	const testURL = `https://practicum.yandex.ru`
	req, err := http.NewRequest(http.MethodPost, fakeServ.URL+"/", strings.NewReader(testURL))
	if err != nil {
		t.Fatal(err)
	}
	rr := httptest.NewRecorder()
	service.URLEncode(rr, req)
	resp := rr.Result()
	defer resp.Body.Close()
	retURL, _ := io.ReadAll(resp.Body)
	lIndex := strings.LastIndex(string(retURL), "/")
	if lIndex < 0 {
		return
	}
	id := retURL[lIndex+1:]
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
			req, err := http.NewRequest(test.method, fakeServ.URL+"/"+string(id), nil)
			if err != nil {
				t.Fatal(err)
			}
			rr = httptest.NewRecorder()
			service.URLDecode(rr, req)
			resp = rr.Result()
			assert.Equal(t, test.status, resp.StatusCode)
			if resp.StatusCode == http.StatusTemporaryRedirect {
				location := resp.Header.Get("Location")
				assert.Equal(t, test.uri, location)
			}
			resp.Body.Close()
		})
	}
	fakeServ.Close()
}
