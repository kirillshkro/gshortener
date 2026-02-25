package shortener

import (
	"compress/gzip"
	"io"
	"net/http"
	"strings"
)

type respCompressWriter struct {
	http.ResponseWriter
}

func HandlerWithCompress(next http.Handler) http.Handler {
	compressFn := func(w http.ResponseWriter, r *http.Request) {
		if !isCompContent(r) {
			next.ServeHTTP(w, r)
			return
		}
		var rBody io.ReadCloser
		compEncoding := r.Header.Get("Accept-Encoding")
		if compEncoding == "gzip" {
			rBody = r.Body
			compReader, err := newCompReader(rBody)
			if err != nil {
				http.Error(w, "Internal error", http.StatusBadRequest)
				return
			}
			r.Body = compReader
			defer compReader.Close()
			w.Header().Set("Content-Type", "application/json")
			next.ServeHTTP(newRespCompressWriter(w), r)
		}
	}
	return http.HandlerFunc(compressFn)
}

func (w *respCompressWriter) Write(b []byte) (int, error) {
	var (
		wr  io.WriteCloser
		err error
		n   int
	)
	wr = gzip.NewWriter(w.ResponseWriter)
	defer wr.Close()
	if n, err = wr.Write(b); err != nil {
		return 0, err
	}

	return n, err
}

func (w *respCompressWriter) WriteHeader(statusCode int) {
	if statusCode == http.StatusCreated {
		w.Header().Set("Content-Encoding", "gzip")
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

func newRespCompressWriter(resp http.ResponseWriter) *respCompressWriter {
	return &respCompressWriter{resp}
}

func isCompContent(r *http.Request) bool {
	if isJSON(r) || isText(r) {
		return true
	}
	return false
}

func isJSON(r *http.Request) bool {
	cType := r.Header.Get("Content-Type")
	return strings.Contains(cType, "application/json")
}

func isText(r *http.Request) bool {
	cType := r.Header.Get("Content-Type")
	return strings.Contains(cType, "plain/text")
}
