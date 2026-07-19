package shortener

import (
	"context"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kirillshkro/gshortener/internal/config/auth"
	"github.com/kirillshkro/gshortener/internal/handler/shortener/claims"
	"github.com/kirillshkro/gshortener/internal/types"
	"github.com/stretchr/testify/suite"
)

type AuthMiddlewareTestSuite struct {
	suite.Suite
	service  *Service
	authCfg  *auth.AuthConfig
}

func (s *AuthMiddlewareTestSuite) SetupSuite() {
	s.authCfg = auth.NewAuthConfig()
	s.service = NewService()
}

func (s *AuthMiddlewareTestSuite) TearDownSuite() {
}

func (s *AuthMiddlewareTestSuite) Test_AuthMiddleware_NoCookie() {
	var handlerCalled bool

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		userID, ok := r.Context().Value(types.UserID).(string)
		s.Assert().False(ok)
		s.Assert().Empty(userID)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	recorder := httptest.NewRecorder()

	authenticatedHandler := s.service.AuthMiddleware(next)
	authenticatedHandler.ServeHTTP(recorder, req)

	s.Assert().True(handlerCalled)
	s.Assert().Equal(http.StatusOK, recorder.Code)
}

func (s *AuthMiddlewareTestSuite) Test_AuthMiddleware_InvalidCookie() {
	var handlerCalled bool

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		userID, ok := r.Context().Value(types.UserID).(string)
		s.Assert().False(ok)
		s.Assert().Empty(userID)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  "auth_cookie",
		Value: "a.b.c",
	})
	recorder := httptest.NewRecorder()

	authenticatedHandler := s.service.AuthMiddleware(next)
	authenticatedHandler.ServeHTTP(recorder, req)

	s.Assert().False(handlerCalled)
	s.Assert().Equal(http.StatusUnauthorized, recorder.Code)
}

func (s *AuthMiddlewareTestSuite) Test_AuthMiddleware_ValidCookie() {
	authUser := claims.NewAuthUser(s.authCfg)
	token, err := authUser.Token()
	s.Assert().NoError(err)
	s.Assert().NotEmpty(token)

	var handlerCalled bool
	var receivedUserID string

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		userID, ok := r.Context().Value(types.UserID).(string)
		s.Assert().True(ok)
		receivedUserID = userID
		s.Assert().NotEmpty(userID)
		s.Assert().Equal(authUser.UserID, userID)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  "auth_cookie",
		Value: token,
	})
	recorder := httptest.NewRecorder()

	authenticatedHandler := s.service.AuthMiddleware(next)
	authenticatedHandler.ServeHTTP(recorder, req)

	s.Assert().True(handlerCalled)
	s.Assert().Equal(http.StatusOK, recorder.Code)
	s.Assert().Equal(authUser.UserID, receivedUserID)
}

func (s *AuthMiddlewareTestSuite) Test_AuthMiddleware_EmptyCookieValue() {
	var handlerCalled bool

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		userID, ok := r.Context().Value(types.UserID).(string)
		s.Assert().False(ok)
		s.Assert().Empty(userID)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  "auth_cookie",
		Value: "",
	})
	recorder := httptest.NewRecorder()

	authenticatedHandler := s.service.AuthMiddleware(next)
	authenticatedHandler.ServeHTTP(recorder, req)

	s.Assert().True(handlerCalled)
	s.Assert().Equal(http.StatusOK, recorder.Code)
}

func (s *AuthMiddlewareTestSuite) Test_AuthMiddleware_CustomContextValue() {
	authUser := claims.NewAuthUser(s.authCfg)
	token, err := authUser.Token()
	s.Assert().NoError(err)

	var handlerCalled bool

	next := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		handlerCalled = true
		ctxUserID, ok := r.Context().Value(types.UserID).(string)
		s.Assert().True(ok)
		s.Assert().Equal(authUser.UserID, ctxUserID)

		customKey := "custom_key"
		customValue := "custom_value"
		r = r.WithContext(context.WithValue(r.Context(), customKey, customValue))
		w.Header().Set("X-Custom", customValue)
	})

	req := httptest.NewRequest(http.MethodGet, "/test", nil)
	req.AddCookie(&http.Cookie{
		Name:  "auth_cookie",
		Value: token,
	})
	recorder := httptest.NewRecorder()

	authenticatedHandler := s.service.AuthMiddleware(next)
	authenticatedHandler.ServeHTTP(recorder, req)

	s.Assert().True(handlerCalled)
	s.Assert().Equal(http.StatusOK, recorder.Code)
	s.Assert().Equal("custom_value", recorder.Header().Get("X-Custom"))
}

func TestAuthMiddleware(t *testing.T) {
	suite.Run(t, new(AuthMiddlewareTestSuite))
}
