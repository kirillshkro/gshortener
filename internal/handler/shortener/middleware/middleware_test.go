package middleware

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"
	"time"

	"github.com/kirillshkro/gshortener/internal/handler/shortener"
	"github.com/kirillshkro/gshortener/internal/types"
	"github.com/stretchr/testify/assert"
)

type logCaptureWriter struct {
	strings.Builder
}

func (l *logCaptureWriter) Write(p []byte) (n int, err error) {
	return l.Builder.Write(p)
}

func TestHandlerWithLog(t *testing.T) {
	service := shortener.NewService()

	wrapped := HandlerWithLog(DecodeHandler(service))

	req := httptest.NewRequest("GET", "/test", nil)
	w := httptest.NewRecorder()

	start := time.Now()
	wrapped.ServeHTTP(w, req)
	elapsed := time.Since(start)
	resp := w.Result()
	defer resp.Body.Close()

	assert.Equal(t, http.StatusTemporaryRedirect, resp.StatusCode)
	assert.NotZero(t, elapsed)

	reqData := types.RequestData{
		URL: "https://weather.google.com",
	}

	body := bytes.NewBuffer(nil)
	if err := json.NewEncoder(body).Encode(reqData); err != nil {
		t.Fatal(err)
	}
	req = httptest.NewRequest(http.MethodPost, "/api/shorten", body)
	w = httptest.NewRecorder()

	start = time.Now()

	wrapped = HandlerWithLog(EncodeHandler(service))
	wrapped.ServeHTTP(w, req)
	elapsed = time.Since(start)
	resp = w.Result()
	defer resp.Body.Close()

	if assert.Equal(t, http.StatusCreated, resp.StatusCode) {
		assert.Greater(t, elapsed, time.Millisecond*0)
	}
}
