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
	"github.com/kirillshkro/gshortener/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ServiceTestsSuite struct {
	suite.Suite
	service *Service
	router  *mux.Router
	server  *httptest.Server
}

func (s *ServiceTestsSuite) SetupSuite() {
	s.router = mux.NewRouter()
	s.server = httptest.NewServer(s.router)
	s.service = NewServiceWithAddrWithAddrShortener(types.RawURL(s.server.URL), types.ShortURL(s.server.URL))
	s.router.HandleFunc("/", s.service.URLEncode).Methods(http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch)
	s.router.HandleFunc("/{id}", s.service.URLDecode).Methods(http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch)
}

func (s *ServiceTestsSuite) TearDownSuite() {
	s.server.Close()
	s.service.Stor.Close()
}

func (s *ServiceTestsSuite) Test_URLEncode() {
	const testData = "https://practicum.yandex.ru"
	tests := []struct {
		name   string
		method string
		uri    string
		hash   types.ShortURL
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
		s.T().Run(test.name, func(t *testing.T) {
			body, err := json.Marshal(test.uri)
			if err != nil {
				t.Fatal(err)
			}
			req, err := http.NewRequest(test.method, s.server.URL+"/", strings.NewReader(string(body)))
			if err != nil {
				t.Fatal(err)
			}
			client := s.server.Client()
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

func (s *ServiceTestsSuite) Test_URLDecode() {
	const testURL = `https://practicum.yandex.ru`
	id := Hashing([]byte(testURL))
	if err := s.service.Stor.SetData(types.URLData{
		ShortURL:    types.ShortURL(id),
		OriginalURL: types.RawURL(testURL),
	}); err != nil {
		s.T().Fatal(err)
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
		s.T().Run(test.name, func(t *testing.T) {
			req, err := http.NewRequest(test.method, s.server.URL+"/"+string(id), nil)
			if err != nil {
				t.Fatal(err)
			}
			rr := httptest.NewRecorder()
			s.service.URLDecode(rr, req)
			resp := rr.Result()
			defer resp.Body.Close()
			assert.Equal(t, test.status, resp.StatusCode)
			if resp.StatusCode == http.StatusTemporaryRedirect {
				location := resp.Header.Get("Location")
				assert.Equal(t, test.uri, location)
			}
		})
	}
}

func (s *ServiceTestsSuite) Test_CreateShortURL() {
	var (
		mockResp bytes.Buffer
		mockReq  types.RequestData
	)
	testURL := types.RawURL("https://weather.yandex.ru")
	mockReq.URL = testURL
	testBody, err := json.Marshal(mockReq)
	if err != nil {
		s.T().Fatal(err)
	}
	id := Hashing([]byte(testURL))
	shortedURL := types.ShortURL(s.server.URL) + "/" + id
	respData := types.ResponseData{
		Result: shortedURL,
	}
	if err := json.NewEncoder(&mockResp).Encode(respData); err != nil {
		s.T().Fatal(err)
	}
	tests := []struct {
		name   string
		method string
		body   []byte
		status int
	}{
		{
			name:   "Normal POST request",
			method: http.MethodPost,
			body:   testBody,
			status: http.StatusCreated,
		},
		{
			name:   "Empty body POST request",
			method: http.MethodPost,
			body:   []byte(""),
			status: http.StatusBadRequest,
		},
		{
			name:   "Wrong GET request",
			method: http.MethodGet,
			body:   []byte(""),
			status: http.StatusBadRequest,
		},
	}
	for _, test := range tests {
		s.T().Run(test.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(test.method, s.server.URL+"/api/shortlen", bytes.NewBuffer(test.body))
			s.service.CreateShortURL(rr, req)
			resp := rr.Result()
			defer resp.Body.Close()
			assert.Equalf(t, test.status, resp.StatusCode, "Test failed by status code expected: %d, actual: %d\n", test.status, resp.StatusCode)
			if resp.StatusCode == http.StatusCreated {
				rBody, _ := io.ReadAll(resp.Body)
				assert.JSONEqf(t, string(rBody), mockResp.String(), "Test failed by body expected: %v, actual: %v\n", mockResp.String(), string(rBody))
			}
		})
	}
}

func Test_Main(t *testing.T) {
	suite.Run(t, new(ServiceTestsSuite))
}
