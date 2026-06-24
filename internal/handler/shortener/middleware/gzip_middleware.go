// Package middleware provides HTTP middleware
package middleware

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

// HandlerWithGzip is an HTTP middleware that handles gzip compression for both
// request bodies and response bodies based on the Accept-Encoding and Content-Encoding headers
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

// Header returns the header map that will be sent by WriteHeader
func (c *compWriter) Header() http.Header {
	return c.w.Header()
}

// Write writes data to the gzip writer
func (c *compWriter) Write(b []byte) (int, error) {
	return c.zw.Write(b)
}

// WriteHeader writes the HTTP response header and sets the Content-Encoding header
// to "gzip" if the status code is less than 300
func (c *compWriter) WriteHeader(statusCode int) {
	if statusCode < 300 {
		c.w.Header().Set("Content-Encoding", "gzip")
	}
	c.w.WriteHeader(statusCode)
}

// Close closes the gzip writer
func (c *compWriter) Close() error {
	return c.zw.Close()
}

func newCompWriter(w http.ResponseWriter) *compWriter {
	return &compWriter{
		w:  w,
		zw: gzip.NewWriter(w),
	}
}
