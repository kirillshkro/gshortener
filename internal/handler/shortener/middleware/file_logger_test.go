package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/gorilla/mux"
	"github.com/kirillshkro/gshortener/internal/handler/shortener"
	"github.com/kirillshkro/gshortener/internal/types"
	"github.com/stretchr/testify/suite"
)

type FileLoggerSuite struct {
	suite.Suite
	serv   *shortener.Service
	router *mux.Router
	server *httptest.Server
}

func (s *FileLoggerSuite) SetupSuite() {
	os.Setenv("FILE_STORAGE_PATH", "./shortener.json")
	s.router = mux.NewRouter()
	s.server = httptest.NewServer(s.router)
	s.serv = shortener.NewServiceWithAddrWithAddrShortener(types.RawURL(s.server.URL), types.ShortURL(s.server.URL))
	s.router.HandleFunc("/api/shorten", s.serv.CreateShortURL).Methods(http.MethodGet, http.MethodPost, http.MethodPut, http.MethodPatch)
}

func (s *FileLoggerSuite) TestFileLog() {
	rData := shortener.RequestData{
		URL: "https://weather.yandex.ru",
	}
	var (
		buf      bytes.Buffer
		respData shortener.ResponseData
	)
	if err := json.NewEncoder(&buf).Encode(rData); err != nil {
		s.Fail(err.Error())
	}
	body := bytes.NewReader(buf.Bytes())
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, s.server.URL+"/api/shorten", body)
	HandlerLogDatabase(shortener.CreateShortURLHandler(s.serv)).ServeHTTP(rr, req)
	resp := rr.Result()
	defer resp.Body.Close()
	if !s.Equal(http.StatusCreated, resp.StatusCode) {
		s.Fail("wrong status code")
	}
	fps, found := os.LookupEnv("FILE_STORAGE_PATH")
	if !found {
		s.Fail("FILE_STORAGE_PATH not found")
	}
	s.Assert().NotEmpty(fps)
	if err := json.NewDecoder(resp.Body).Decode(&respData); err != nil {
		s.Fail(err.Error())
	}
	if s.NotEmpty(respData.Result) {
		f, err := os.Open(fps)
		if err != nil {
			s.Fail(err.Error())
		}
		defer f.Close()
		stat, err := f.Stat()
		if err != nil {
			s.Fail(err.Error())
		}
		s.Assert().Greater(stat.Size(), int64(0))
	}
}

func (s *FileLoggerSuite) TearDownSuite() {
	s.server.Close()
}

func TestMain(t *testing.T) {
	suite.Run(t, new(FileLoggerSuite))
}
