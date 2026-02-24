package shortener

import (
	"bytes"
	"compress/gzip"
	"encoding/json"
	"io"
	"net/http"
	"net/http/httptest"
	"testing"

	"github.com/kirillshkro/gshortener/internal/types"
	"github.com/stretchr/testify/assert"
)

func Test_HandlerWithCompressGzipPlain(t *testing.T) {
	var (
		buffer bytes.Buffer
		outBuf bytes.Buffer
		tb     bytes.Buffer
	)
	ts := new(TestCompressSuite)
	ts.setUp()
	ts.service = NewServiceWithAddrWithAddrShortener(types.RawURL(ts.server.URL), types.ShortURL(ts.server.URL))
	defer ts.tearDown()

	testCases := []struct {
		name             string
		compMethod       string
		expectedResponse string
	}{
		{
			name:             "Test with gzip compression",
			compMethod:       "gzip",
			expectedResponse: "Content-Encoding: gzip",
		},
	}

	URL := "https://weather.yandex.ru"
	if err := json.NewEncoder(&tb).Encode(URL); err != nil {
		t.Fatal(err)
	}
	hs := Hashing(tb.Bytes())
	extectedStr := ts.server.URL + "/" + hs
	if err := json.NewEncoder(&buffer).Encode(URL); err != nil {
		t.Fatal(err)
	}
	gz := gzip.NewWriter(&outBuf)
	if _, err := gz.Write(buffer.Bytes()); err != nil {
		t.Fatal(err)
	}
	gz.Close()
	for _, test := range testCases {
		t.Run(test.name, func(t *testing.T) {
			req := httptest.NewRequest(http.MethodPost, ts.server.URL+"/", &outBuf)
			req.Header.Set("Accept-Encoding", test.compMethod)
			req.Header.Set("Content-Type", "application/json")
			rr := httptest.NewRecorder()

			HandlerWithCompress(EncodeHandler(ts.service)).ServeHTTP(rr, req)
			resp := rr.Result()
			defer resp.Body.Close()

			var unpacked []byte

			unz, err := gzip.NewReader(resp.Body)
			if err != nil {
				t.Fatal(err)
			}
			defer unz.Close()

			if unpacked, err = io.ReadAll(unz); err != nil {
				t.Fatal(err)
			}

			assert.Equal(t, test.compMethod, resp.Header.Get("Content-Encoding"))
			assert.Equal(t, http.StatusCreated, resp.StatusCode)
			assert.Equal(t, extectedStr, string(unpacked))
		})
	}
}
