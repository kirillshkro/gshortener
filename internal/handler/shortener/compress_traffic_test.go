package shortener

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
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

func Test_HandlerWithCompress(t *testing.T) {
	var (
		testBuffer bytes.Buffer
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
		{
			name:             "Compress with deflate",
			compMethod:       "deflate",
			expectedResponse: "Accept-Encoding: deflate",
		},
	}

	testData := RequestData{
		URL: genContent("https://weather.yandex.ru"),
	}

	if err := json.NewEncoder(&testBuffer).Encode(testData); err != nil {
		t.Fatal(err)
	}

	actualSize := len(testBuffer.String())

	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		w.WriteHeader(http.StatusCreated)
		if _, err := w.Write(testBuffer.Bytes()); err != nil {
			t.Fatal(err)
		}
	})

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, ts.server.URL+"/api/shorten", &testBuffer)
			req.Header.Set("Content-Encoding", test.compMethod)
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			HandlerWithCompress(fn).ServeHTTP(rr, req)

			resp := rr.Result()
			defer resp.Body.Close()

			assert.Equal(t, http.StatusCreated, resp.StatusCode)
			assert.Equal(t, test.compMethod, resp.Header.Get("Accept-Encoding"))

			compressedBody, err := io.ReadAll(resp.Body)
			if err != nil {
				t.Fatal(err)
			}
			assert.Less(t, len(compressedBody), actualSize, "Compressed body should be smaller than original\n")
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

func genContent(str string) string {
	var (
		sb strings.Builder
	)

	for range 1000 {
		sb.WriteString(str)
	}
	return sb.String()
}
