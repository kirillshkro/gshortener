package shortener

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

func HandlerWithGzip(next http.Handler) http.Handler {
	fn := func(w http.ResponseWriter, r *http.Request) {
		zw := w
		isCompressed := r.Header.Get("Accept-Encoding")
		isGzipped := strings.Contains(isCompressed, "gzip")
		if isGzipped {
			cw := newCompWriter(w)
			zw = cw
			defer cw.zw.Close()
		}

		encoding := r.Header.Get("Content-encoding")
		respGzip := strings.Contains(encoding, "gzip")

		if respGzip {
			gzr, err := newCompReader(r.Body)
			if err != nil {
				http.Error(w, "unkwown server error: "+err.Error(), http.StatusInternalServerError)
				return
			}
			r.Body = gzr
			defer gzr.Close()
		}
		next.ServeHTTP(zw, r)
	}
	return http.HandlerFunc(fn)
}

type compWriter struct {
	w  http.ResponseWriter
	zw io.WriteCloser
}

func (c *compWriter) Header() http.Header {
	return c.w.Header()
}

func (c *compWriter) Write(b []byte) (int, error) {
	return c.zw.Write(b)
}

func (c *compWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

func (c *compWriter) Close() error {
	return c.zw.Close()
}

func newCompWriter(w http.ResponseWriter) *compWriter {
	return &compWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}
