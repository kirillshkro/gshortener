package shortener

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kirillshkro/gshortener/internal/types"
	"github.com/stretchr/testify/assert"
)

func Test_CreateShortURL(t *testing.T) {
	var (
		mockResp bytes.Buffer
		mockReq  types.RequestData
	)
	testURL := types.RawURL("https://weather.yandex.ru")
	mockReq.URL = testURL
	testBody, err := json.Marshal(mockReq)
	if err != nil {
		t.Fatal(err)
	}
	id := Hashing([]byte(testURL))
	service, server := setup()
	defer server.Close()
	shortedURL := types.ShortURL(server.URL) + "/" + id
	respData := types.ResponseData{
		Result: shortedURL,
	}
	if err := json.NewEncoder(&mockResp).Encode(respData); err != nil {
		t.Fatal(err)
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
		t.Run(test.name, func(t *testing.T) {
			rr := httptest.NewRecorder()
			req := httptest.NewRequest(test.method, server.URL+"/api/shortlen", bytes.NewBuffer(test.body))
			service.CreateShortURL(rr, req)
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
