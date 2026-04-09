package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"
	"time"

	"github.com/kirillshkro/gshortener/internal/handler/shortener"
	"github.com/kirillshkro/gshortener/internal/types"
	"github.com/kirillshkro/gshortener/pkg/urlgen"
	"github.com/stretchr/testify/suite"
)

type HandlerLogTestSuite struct {
	suite.Suite
	service *shortener.Service
}

func (s *HandlerLogTestSuite) SetupSuite() {
	s.service = shortener.NewService()
}

func (s *HandlerLogTestSuite) TearDownSuite() {
}

func (s *HandlerLogTestSuite) TestHandlerWithLog() {
	wrapped := HandlerWithLog(DecodeHandler(s.service))

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	w := httptest.NewRecorder()

	start := time.Now()
	wrapped.ServeHTTP(w, req)
	elapsed := time.Since(start)
	resp1 := w.Result()
	defer resp1.Body.Close()

	s.Assert().Equal(http.StatusTemporaryRedirect, resp1.StatusCode)
	s.Assert().Positive(elapsed)

	reqData := types.RequestData{
		URL: types.RawURL(urlgen.GenerateURL("https://base.com")),
	}

	body := bytes.NewBuffer(nil)
	if err := json.NewEncoder(body).Encode(reqData); err != nil {
		s.T().Fatal(err)
	}
	req = httptest.NewRequest(http.MethodPost, "/api/shorten", body)
	w2 := httptest.NewRecorder()

	start = time.Now()

	wrapped = HandlerWithLog(EncodeHandler(s.service))
	wrapped.ServeHTTP(w2, req)
	elapsed = time.Since(start)

	s.Assert().Condition(func() bool {
		return w2.Code == http.StatusCreated || w2.Code == http.StatusConflict
	}, "expected status code 201 or 409, got %d", w2.Code)
	s.Assert().Greater(elapsed, time.Microsecond*0)
}

func TestMiddleware_Main(t *testing.T) {
	suite.Run(t, new(HandlerLogTestSuite))
}
