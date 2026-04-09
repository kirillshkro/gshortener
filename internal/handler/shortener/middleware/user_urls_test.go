package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kirillshkro/gshortener/internal/config/auth"
	"github.com/kirillshkro/gshortener/internal/handler/shortener"
	"github.com/kirillshkro/gshortener/internal/handler/shortener/claims"
	"github.com/kirillshkro/gshortener/internal/types"
	"github.com/kirillshkro/gshortener/pkg/urlgen"
	"github.com/stretchr/testify/suite"
)

type TestURLSuite struct {
	suite.Suite
	token   string
	user    claims.AuthUser
	service *shortener.Service
}

func (s *TestURLSuite) SetupSuite() {
	cfg := auth.NewAuthConfig()
	s.user = *claims.NewAuthUser(cfg)
	s.token, _ = s.user.Token()
	s.service = shortener.NewService()
}

func (s *TestURLSuite) TearDownSuite() {
}

func (s *TestURLSuite) Test_GetUserURLs() {
	testURL := urlgen.GenerateURL("tttp://abracadabra")
	postReq := httptest.NewRequest(http.MethodPost, "/", bytes.NewBufferString(testURL))
	rr := httptest.NewRecorder()
	s.service.URLEncode(rr, postReq)
	resp := rr.Result()
	defer resp.Body.Close()
	cookie := resp.Header.Get("Set-Cookie")

	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	req.Header.Set("Cookie", cookie)
	rr = httptest.NewRecorder()
	s.service.GetUserURLs(rr, req)
	resp = rr.Result()
	defer resp.Body.Close()
	var buffer []types.UserURL
	err := json.NewDecoder(resp.Body).Decode(&buffer)
	s.Assert().NoError(err)
	s.Assert().Equal(http.StatusOK, resp.StatusCode)
	s.Assert().Equal("application/json", resp.Header.Get("Content-Type"))
	s.Assert().NotEmptyf(resp.Cookies(), "Cookies not found")
}

func TestMain(t *testing.T) {
	suite.Run(t, new(TestURLSuite))
}
