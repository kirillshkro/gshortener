package shortener

import (
	"fmt"
	"net/http"
	"os"
	"time"

	"github.com/rs/zerolog"
)

type respWriter struct {
	http.ResponseWriter
	statusCode int
	size       int
}

func HandlerWithLog(h http.Handler) http.Handler {
	logger := zerolog.New(os.Stderr)
	infoFn := func(resp http.ResponseWriter, req *http.Request) {
		writer := newRespWriter(resp)
		uri := req.RequestURI
		method := req.Method
		tRec := time.Now()
		h.ServeHTTP(resp, req)
		interval := time.Since(tRec)
		logger.Info().Msg(fmt.Sprintf("URI request: %s\t, method: %s\t, time: %v\n", uri, method, interval))
		size := writer.size
		statusCode := writer.statusCode
		logger.Info().Msg(fmt.Sprintf("Content length %d\t, status code: %d\n", size, statusCode))
	}
	return http.HandlerFunc(infoFn)
}

func newRespWriter(resp http.ResponseWriter) *respWriter {
	return &respWriter{resp, http.StatusOK, 0}
}

func (w *respWriter) Write(b []byte) (int, error) {
	n, err := w.ResponseWriter.Write(b)
	w.size += n
	return n, err
}

func (w *respWriter) WriteHeader(statusCode int) {
	w.statusCode = statusCode
	w.ResponseWriter.WriteHeader(statusCode)
}

func EncodeHandler(s IService) http.Handler {
	return http.HandlerFunc(s.URLEncode)
}

func DecodeHandler(s IService) http.Handler {
	return http.HandlerFunc(s.URLDecode)
}

func JSONEncodeHandler(s JSONEncoder) http.Handler {
	return http.HandlerFunc(s.JSONEncode)
}
