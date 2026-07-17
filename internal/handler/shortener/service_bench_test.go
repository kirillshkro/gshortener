package shortener

import (
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kirillshkro/gshortener/pkg/urlgen"
)

func Benchmark_URLEncode(b *testing.B) {
	s := new(ServiceTestsSuite)
	s.SetT(&testing.T{})
	s.SetupSuite()
	url := urlgen.GenerateURL("http://basedurl")
	req := httptest.NewRequest(http.MethodPost, url, nil)
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
	shortedUrl := Hashing([]byte(url))
	req := httptest.NewRequest(http.MethodGet, "/"+string(shortedUrl), nil)
	rr := httptest.NewRecorder()
	b.ResetTimer()
	for b.Loop() {
		s.service.URLDecode(rr, req)
	}
}
