package shortener

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kirillshkro/gshortener/internal/handler/shortener/claims"
	"github.com/kirillshkro/gshortener/internal/repository/storage"
	"github.com/kirillshkro/gshortener/internal/types"
	"github.com/kirillshkro/gshortener/pkg/urlgen"
	"github.com/stretchr/testify/suite"
)

type UserURLsTestSuite struct {
	suite.Suite
	service *Service
	userID  string
	storage *storage.MemoryStorage
}

func (s *UserURLsTestSuite) SetupSuite() {
	s.storage = storage.NewMemoryStorage()
	s.service = NewService()
	s.service.Stor = s.storage

	s.userID = "test-user-id"
}

func (s *UserURLsTestSuite) TearDownSuite() {
}

func (s *UserURLsTestSuite) Test_GetUserURLs_Successful() {
	url := urlgen.GenerateURL("https://test.com")

	req := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(url)))
	recorder := httptest.NewRecorder()
	s.service.URLEncode(recorder, req)
	resp := recorder.Result()
	s.Assert().Equal(http.StatusCreated, resp.StatusCode)

	cookie := resp.Cookies()[0]
	token := cookie.Value
	userID, err := claims.GetUserID(token)
	s.Assert().NoError(err)

	req = httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	req = req.WithContext(context.WithValue(req.Context(), types.UserID, userID))
	req.AddCookie(cookie)
	recorder = httptest.NewRecorder()
	s.service.GetUserURLs(recorder, req)
	resp = recorder.Result()
	defer resp.Body.Close()

	s.Assert().Equal(http.StatusOK, resp.StatusCode)

	var userURLs []types.UserURL
	err = json.NewDecoder(resp.Body).Decode(&userURLs)
	s.Assert().NoError(err)
	s.Assert().Len(userURLs, 1)
	s.Assert().Contains(userURLs[0].ShortURL, string(s.service.ResultAddr))
	s.Assert().Equal(url, userURLs[0].OriginalURL)
}

func (s *UserURLsTestSuite) Test_GetUserURLs_EmptyList() {
	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	recorder := httptest.NewRecorder()
	s.service.GetUserURLs(recorder, req)
	resp := recorder.Result()
	defer resp.Body.Close()

	s.Assert().Equal(http.StatusNoContent, resp.StatusCode)
}

func (s *UserURLsTestSuite) Test_GetUserURLs_MultipleURLs() {
	url1 := urlgen.GenerateURL("https://test1.com")
	url2 := urlgen.GenerateURL("https://test2.com")

	req1 := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(url1)))
	recorder1 := httptest.NewRecorder()
	s.service.URLEncode(recorder1, req1)
	resp1 := recorder1.Result()
	s.Assert().Equal(http.StatusCreated, resp1.StatusCode)

	cookie := resp1.Cookies()[0]
	token := cookie.Value
	userID, err := claims.GetUserID(token)
	s.Assert().NoError(err)

	req2 := httptest.NewRequest(http.MethodPost, "/", bytes.NewReader([]byte(url2)))
	req2.AddCookie(cookie)
	recorder2 := httptest.NewRecorder()
	s.service.URLEncode(recorder2, req2)
	resp2 := recorder2.Result()
	s.Assert().Equal(http.StatusCreated, resp2.StatusCode)

	req3 := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	req3.AddCookie(cookie)
	req3 = req3.WithContext(context.WithValue(req3.Context(), types.UserID, userID))
	recorder3 := httptest.NewRecorder()
	s.service.GetUserURLs(recorder3, req3)
	resp3 := recorder3.Result()
	defer resp3.Body.Close()

	s.Assert().Equal(http.StatusOK, resp3.StatusCode)

	var userURLs []types.UserURL
	err = json.NewDecoder(resp3.Body).Decode(&userURLs)
	s.Assert().NoError(err)
	s.Assert().Len(userURLs, 2)
}

func (s *UserURLsTestSuite) Test_GetUserURLs_NoUserID() {
	req := httptest.NewRequest(http.MethodGet, "/api/user/urls", nil)
	recorder := httptest.NewRecorder()
	s.service.GetUserURLs(recorder, req)
	resp := recorder.Result()
	defer resp.Body.Close()

	s.Assert().Equal(http.StatusNoContent, resp.StatusCode)
}

func TestUserURLs(t *testing.T) {
	suite.Run(t, new(UserURLsTestSuite))
}
