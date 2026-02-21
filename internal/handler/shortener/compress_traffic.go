package shortener

import (
	"compress/gzip"
	"compress/zlib"
	"io"
	"net/http"
)

type respCompressWriter struct {
	http.ResponseWriter
	EncodingType string
}

func HandlerWithCompress(next http.Handler) http.Handler {
	compressFn := func(w http.ResponseWriter, r *http.Request) {
		if isCompContent(r) {
			compEncoding := r.Header.Get("Content-Encoding")
			switch compEncoding {
			case "gzip":
				w.Header().Set("Accept-Encoding", "gzip")
				next.ServeHTTP(newRespCompressWriter(w, compEncoding), r)
			case "deflate":
				w.Header().Set("Accept-Encoding", "deflate")
				next.ServeHTTP(newRespCompressWriter(w, compEncoding), r)
			default:
				next.ServeHTTP(w, r)
			}
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
	switch w.EncodingType {
	case "gzip":
		wr, err = gzip.NewWriterLevel(w.ResponseWriter, gzip.DefaultCompression)
		if err != nil {
			return 0, err
		}
	case "deflate":
		wr, err = zlib.NewWriterLevel(w.ResponseWriter, zlib.DefaultCompression)
		if err != nil {
			return 0, err
		}
	}
	defer wr.Close()
	if n, err = wr.Write(b); err != nil {
		return 0, err
	}

	return n, err
}

func (w *respCompressWriter) WriteHeader(statusCode int) {
	if statusCode == http.StatusCreated {
		w.Header().Set("Accept-Encoding", w.EncodingType)
	}
	w.ResponseWriter.WriteHeader(statusCode)
}

func newRespCompressWriter(resp http.ResponseWriter, compType string) *respCompressWriter {
	return &respCompressWriter{resp, compType}
}

func isCompContent(r *http.Request) bool {
	cType := r.Header.Get("Content-Type")
	if cType == "text/html" || cType == "application/json" {
		return true
	}
	return false
}
