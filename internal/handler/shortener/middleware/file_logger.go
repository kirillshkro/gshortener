package middleware

import (
	"bytes"
	"encoding/json"
	"io"
	"net/http"
	"strings"
	"sync/atomic"

	"github.com/kirillshkro/gshortener/internal/config"
	"github.com/kirillshkro/gshortener/internal/handler/shortener"
	"github.com/kirillshkro/gshortener/internal/repository/storage"
	"github.com/kirillshkro/gshortener/internal/types"
)

type infoWriter struct {
	http.ResponseWriter
	respBody *bytes.Buffer
	counter  int64
	fstor    *storage.FileStorage
}

func (w *infoWriter) Write(b []byte) (int, error) {
	return w.ResponseWriter.Write(b)
}

func HandlerLogDatabase(next http.Handler) http.Handler {
	var seq uint64

	fn := http.HandlerFunc(func(w http.ResponseWriter, r *http.Request) {
		if r.Method != http.MethodPost {
			next.ServeHTTP(w, r)
			return
		}
		var (
			rdata    shortener.RequestData
			respData shortener.ResponseData
		)

		buf, err := io.ReadAll(r.Body)
		if err != nil {
			http.Error(w, "unkwown server error: "+err.Error(), http.StatusInternalServerError)
			return
		}
		r.Body = io.NopCloser(bytes.NewBuffer(buf))

		if err := json.NewDecoder(bytes.NewReader(buf)).Decode(&rdata); err != nil {
			http.Error(w, "cannot decode request: "+err.Error(), http.StatusBadRequest)
			return
		}

		infoWriter := newInfoWriter()
		next.ServeHTTP(infoWriter, r)

		if err = json.NewDecoder(infoWriter.respBody).Decode(&respData); err != nil {
			http.Error(w, "cannot decode response: "+err.Error(), http.StatusInternalServerError)
			return
		}

		shortURL := respData.Result
		posIndex := strings.LastIndex(shortURL, "/")
		if posIndex == -1 {
			http.Error(w, "cannot parse short url", http.StatusInternalServerError)
			return
		}

		id := shortURL[posIndex+1:]

		atomic.AddUint64(&seq, 1)

		if err = infoWriter.fstor.SetData(types.ShortURL(id), types.RawURL(rdata.URL)); err != nil {
			http.Error(w, "cannot save data to file: "+err.Error(), http.StatusInternalServerError)
			return
		}
	})
	return http.HandlerFunc(fn)
}

func newInfoWriter() *infoWriter {
	cfg := config.GetConfig()
	fStorage, err := storage.GetFileStorage(cfg.FileDb)
	if err != nil {
		return nil
	}
	counter, err := fStorage.GetCounter()
	if err != nil {
		return nil
	}
	iw := &infoWriter{
		respBody: &bytes.Buffer{},
		counter:  counter,
		fstor:    fStorage,
	}
	return iw
}
