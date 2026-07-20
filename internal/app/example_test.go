// Package shortener_test provides test cases for the shortener API endpoints.
package app

import (
	"bytes"
	"context"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kirillshkro/gshortener/internal/handler/shortener"
	"github.com/kirillshkro/gshortener/internal/model"
	"github.com/kirillshkro/gshortener/internal/types"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/suite"
)

type ExamplesSuite struct {
	suite.Suite
	service *shortener.Service
}

func (s *ExamplesSuite) SetupSuite() {
	s.service = shortener.NewServiceWithAddr("http://localhost:8080")
}

// ExampleCreateShortURL demonstrates how to create a short URL using the CreateShortURL method of the Service struct.
func (s *ExamplesSuite) ExampleCreateShortURL() {
	req := httptest.NewRequest(http.MethodPost, "/create", bytes.NewBufferString(`{"url":"https://example.com"}`))
	req.Header.Set("Content-Type", "application/json")

	// Act
	recorder := httptest.NewRecorder()
	s.service.CreateShortURL(recorder, req)

	// Assert
	s.Assert().Equal(http.StatusCreated, recorder.Code)
	var responseData types.ResponseData
	err := json.Unmarshal(recorder.Body.Bytes(), &responseData)
	s.Assert().NoError(err)
	s.Assert().NotEmpty(responseData.Result)
}

// ExampleDecodeURL demonstrates how to decode a URL using the DecodeURL method of the Service struct.
func (s *ExamplesSuite) ExampleDecodeURL() {
	// Arrange
	hash := shortener.Hashing([]byte("https://example.com"))
	s.service.Stor.Create(model.URLData{
		ShortURL:    types.ShortURL(hash),
		OriginalURL: types.RawURL("https://example.com"),
	})

	req := httptest.NewRequest(http.MethodGet, "/"+string(hash), nil)

	// Act
	recorder := httptest.NewRecorder()
	s.service.URLDecode(recorder, req)

	// Assert
	s.Equal(http.StatusTemporaryRedirect, recorder.Code)
	s.Equal("https://example.com", recorder.Header().Get("Location"))
}

// ExampleBatchCreateShortURL demonstrates how to batch create short URLs using the BatchCreateShortURL method of the Service struct.
func (s *ExamplesSuite) ExampleBatchCreateShortURL() {
	// Arrange
	req := httptest.NewRequest(http.MethodPost, "/batch", bytes.NewBufferString(`[{"correlation_id":"1","original_url":"https://example.com"},{"correlation_id":"2","original_url":"https://another-example.com"}]`))
	req.Header.Set("Content-Type", "application/json")

	// Act
	recorder := httptest.NewRecorder()
	s.service.BatchCreateShortURL(recorder, req)

	// Assert
	s.Equal(http.StatusCreated, recorder.Code)
	var response []types.BatchResponse
	err := json.Unmarshal(recorder.Body.Bytes(), &response)
	s.NoError(err)
	s.Len(response, 2)
	s.NotEmpty(response[0].ShortURL)
	s.NotEmpty(response[1].ShortURL)
}

// ExampleGetUserURLs demonstrates how to get user URLs using the GetUserURLs method of the Service struct.
func (s *ExamplesSuite) ExampleGetUserURLs(t *testing.T) {
	// Arrange
	service := shortener.NewServiceWithAddr("http://localhost:8080")
	hash := shortener.Hashing([]byte("https://example.com"))
	userID := "user123"
	service.Stor.Create(model.URLData{
		ShortURL:    types.ShortURL(hash),
		OriginalURL: types.RawURL("https://example.com"),
		UserUUID:    userID,
	})

	req := httptest.NewRequest(http.MethodGet, "/user_urls", nil)
	ctx := context.WithValue(req.Context(), types.UserID, userID)
	req = req.WithContext(ctx)

	// Act
	recorder := httptest.NewRecorder()
	s.service.GetUserURLs(recorder, req)

	// Assert
	assert.Equal(t, http.StatusOK, recorder.Code)
	var userURLs []types.UserURL
	err := json.Unmarshal(recorder.Body.Bytes(), &userURLs)
	assert.NoError(t, err)
	assert.Len(t, userURLs, 1)
	assert.NotEmpty(t, userURLs[0].ShortURL)
	assert.NotEmpty(t, userURLs[0].OriginalURL)
}

// ExampleDeleteUserURLs demonstrates how to delete user URLs using the DeleteUserURLs method of the Service struct.
func (s *ExamplesSuite) ExampleDeleteUserURLs(t *testing.T) {
	// Arrange
	service := shortener.NewServiceWithAddr("http://localhost:8080")
	hash := shortener.Hashing([]byte("https://example.com"))
	userID := "user123"
	service.Stor.Create(model.URLData{
		ShortURL:    types.ShortURL(hash),
		OriginalURL: types.RawURL("https://example.com"),
		UserUUID:    userID,
	})

	req := httptest.NewRequest(http.MethodPost, "/delete", bytes.NewBufferString(`["`+string(hash)+`"]`))
	req.Header.Set("Content-Type", "application/json")
	ctx := context.WithValue(req.Context(), types.UserID, userID)
	req = req.WithContext(ctx)

	// Act
	recorder := httptest.NewRecorder()
	s.service.DeleteUserURLs(recorder, req)

	// Assert
	assert.Equal(t, http.StatusAccepted, recorder.Code)
}

// ExamplePing demonstrates how to perform a health check using the Ping method of the Service struct.
func (s *ExamplesSuite) ExamplePing(t *testing.T) {
	// Arrange
	req := httptest.NewRequest(http.MethodGet, "/ping", nil)

	// Act
	recorder := httptest.NewRecorder()
	s.service.Ping(recorder, req)

	// Assert
	assert.Equal(t, http.StatusOK, recorder.Code)
}
