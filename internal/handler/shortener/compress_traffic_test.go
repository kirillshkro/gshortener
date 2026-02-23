package shortener

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/gorilla/mux"
	"github.com/kirillshkro/gshortener/internal/types"
	"github.com/stretchr/testify/assert"
)

type TestCompressSuite struct {
	router  *mux.Router
	server  *httptest.Server
	service *Service
}

func Test_HandlerWithCompressGzip(t *testing.T) {
	const (
		testStr = "Aaa"
	)
	var (
		buf    bytes.Buffer
		outBuf bytes.Buffer
	)
	ts := new(TestCompressSuite)
	ts.setUp()
	defer ts.tearDown()

	testCases := []struct {
		name             string
		compMethod       string
		expectedResponse string
	}{
		{
			name:             "Compress with gzip",
			compMethod:       "gzip",
			expectedResponse: "Accept-Encoding: gzip",
		},
	}
	testBuffer := make([]byte, 0)
	testData := RequestData{
		URL: "https://weather.yandex.ru/",
	}

	if err := json.NewEncoder(&buf).Encode(testData); err != nil {
		t.Fatal(err)
	}

	gz := gzip.NewWriter(&outBuf)
	if _, err := gz.Write(buf.Bytes()); err != nil {
		t.Fatal(err)
	}

	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		var rData RequestData
		w.WriteHeader(http.StatusCreated)
		if err := json.NewDecoder(r.Body).Decode(&rData); err != nil {
			http.Error(w, "JSON unpack error", http.StatusLengthRequired)
			return
		}
		if _, err := w.Write([]byte(testStr)); err != nil {
			http.Error(w, "Net I/O error", http.StatusBadRequest)
			return
		}
	})

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, ts.server.URL+"/api/shorten", &outBuf)
			req.Header.Set("Content-Encoding", test.compMethod)
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			HandlerWithCompress(fn).ServeHTTP(rr, req)

			resp := rr.Result()
			defer resp.Body.Close()
			rgz, err := gzip.NewReader(resp.Body)
			if err != nil {
				t.Fatal(err)
			}
			if testBuffer, err = io.ReadAll(rgz); err != nil {
				t.Fatal(err)
			}

			outBuf := string(testBuffer)
			assert.Equal(t, http.StatusCreated, resp.StatusCode)
			assert.Equal(t, test.compMethod, resp.Header.Get("Accept-Encoding"))
			assert.Equal(t, testStr, outBuf)
		})
	}
}

func (s *TestCompressSuite) setUp() {
	s.router = mux.NewRouter()
	s.server = httptest.NewServer(s.router)
	s.service = NewServiceWithAddrWithAddrShortener(types.RawURL("http://localhost:8080"), types.ShortURL("http://localhost:8080"))
	s.router.HandleFunc("/api/shorten", s.service.CreateShortURL).Methods(http.MethodPost)
}

func (s *TestCompressSuite) tearDown() {
	s.server.Close()
}
