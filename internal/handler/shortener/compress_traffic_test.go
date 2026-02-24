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
	var (
		buf        []byte
		outBuf     bytes.Buffer
		mockResp   bytes.Buffer
		err        error
		testBuffer []byte
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
			expectedResponse: "Content-Encoding: gzip",
		},
	}
	testURL := types.RawURL("https://weather.yandex.ru")
	id := Hashing([]byte(testURL))
	shortedURL := types.ShortURL(ts.server.URL + "/" + id)

	testResp := ResponseData{
		Result: string(shortedURL),
	}

	if err = json.NewEncoder(&mockResp).Encode(testResp); err != nil {
		t.Fatal(err)
	}

	testData := RequestData{
		URL: string(testURL),
	}

	buf, err = json.Marshal(testData)
	if err != nil {
		t.Fatal(err)
	}

	gz := gzip.NewWriter(&outBuf)
	if _, err := gz.Write(buf); err != nil {
		t.Fatal(err)
	}
	gz.Close()

	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, ts.server.URL+"/api/shorten", &outBuf)
			req.Header.Set("Accept-Encoding", test.compMethod)
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			HandlerWithCompress(CreateShortURLHandler(ts.service)).ServeHTTP(rr, req)

			resp := rr.Result()
			defer resp.Body.Close()
			rgz, err := gzip.NewReader(resp.Body)
			if err != nil {
				t.Fatal(err)
			}
			defer rgz.Close()
			if testBuffer, err = io.ReadAll(rgz); err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, http.StatusCreated, resp.StatusCode)
			assert.Equal(t, test.compMethod, resp.Header.Get("Content-Encoding"))
			assert.JSONEq(t, mockResp.String(), string(testBuffer))
		})
	}
}

func (s *TestCompressSuite) setUp() {
	s.router = mux.NewRouter()
	s.server = httptest.NewServer(s.router)
	s.service = NewServiceWithAddrWithAddrShortener(types.RawURL(s.server.URL), types.ShortURL(s.server.URL))
	s.router.HandleFunc("/api/shorten", s.service.CreateShortURL).Methods(http.MethodPost)
}

func (s *TestCompressSuite) tearDown() {
	s.server.Close()
}
