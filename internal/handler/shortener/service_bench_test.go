package shortener

import (
	"bytes"
	"encoding/json"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/kirillshkro/gshortener/internal/types"
	"github.com/kirillshkro/gshortener/pkg/urlgen"
)

func Benchmark_URLEncode(b *testing.B) {
	s := new(ServiceTestsSuite)
	s.SetT(&testing.T{})
	s.SetupSuite()
	url := urlgen.GenerateURL("http://basedurl")
	body, err := json.Marshal(url)
	if err != nil {
		b.Fatal(err)
	}
	req := httptest.NewRequest(http.MethodPost, s.server.URL+"/", strings.NewReader(string(body)))
	rr := httptest.NewRecorder()
	b.ResetTimer()
	for b.Loop() {
		s.service.URLEncode(rr, req)
	}
}

func Benchmark_URLDecode(b *testing.B) {
	s := new(ServiceTestsSuite)
	s.SetT(&testing.T{})
	s.SetupSuite()
	url := urlgen.GenerateURL("http://basedurl")
	shortedURL := Hashing([]byte(url))
	req := httptest.NewRequest(http.MethodGet, s.server.URL+"/"+string(shortedURL), nil)
	rr := httptest.NewRecorder()
	b.ResetTimer()
	for b.Loop() {
		s.service.URLDecode(rr, req)
	}
}

func Benchmark_CreateURL(b *testing.B) {
	s := new(ServiceTestsSuite)
	s.SetT(&testing.T{})
	s.SetupSuite()
	url := urlgen.GenerateURL("http://basedurl")
	rr := httptest.NewRecorder()
	b.ResetTimer()
	for b.Loop() {
		rBody := types.RequestData{
			URL: types.RawURL(url),
		}
		reqURL, err := json.Marshal(rBody)
		if err != nil {
			s.T().Fatal(err)
		}
		req := httptest.NewRequest(http.MethodPost, s.server.URL+"/api/shorten", bytes.NewReader(reqURL))
		req.Header.Set("Content-Type", "application/json")
		s.service.CreateShortURL(rr, req)
	}
}
