package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"os"
	"testing"

	"github.com/kirillshkro/gshortener/internal/handler/shortener"
	"github.com/stretchr/testify/suite"
)

type FileLoggerSuite struct {
	suite.Suite
	logger *infoWriter
	serv   *shortener.Service
}

func (s *FileLoggerSuite) SetupSuite() {
	s.logger = newInfoWriter()
	s.serv = shortener.NewService()
}

func (s *FileLoggerSuite) TestFileLog() {
	rData := shortener.RequestData{
		URL: "https://weather.yandex.ru",
	}
	buf := make([]byte, 0)
	if err := json.NewEncoder(bytes.NewBuffer(buf)).Encode(rData); err != nil {
		s.Fail(err.Error())
	}
	body := bytes.NewReader(buf)
	rr := httptest.NewRecorder()
	req := httptest.NewRequest(http.MethodPost, "/", body)
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
}

func (s *FileLoggerSuite) TearDownSuite() {
}

func TestMain(t *testing.T) {
	suite.Run(t, new(FileLoggerSuite))
}
