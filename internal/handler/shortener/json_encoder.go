package shortener

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"
	"time"

	"github.com/kirillshkro/gshortener/internal/model"
	"github.com/kirillshkro/gshortener/internal/types"
)

// JSONEncoder interface defines the contract for creating short URLs
type JSONEncoder interface {
	// CreateShortURL handles the creation of a short URL from a long URL
	CreateShortURL(resp http.ResponseWriter, req *http.Request)
}

// CreateShortURL handles the creation of a short URL from a long URL
// It expects a JSON request with the original URL in the body
// Returns HTTP 201 Created on successful creation or HTTP 409 Conflict if URL already exists
func (s Service) CreateShortURL(resp http.ResponseWriter, req *http.Request) {
	var (
		data     types.RequestData
		respData types.ResponseData
	)
	if req.Method != http.MethodPost {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}
	if err := json.NewDecoder(req.Body).Decode(&data); err != nil {
		log.Println("cannot decode request: ", err.Error())
		resp.WriteHeader(http.StatusBadRequest)
		return
	}
	id := Hashing([]byte(data.URL))
	respData.Result = s.ResultAddr + "/" + id
	s.logger.Info("URL: " + string(s.ResultAddr))
	if err := s.Stor.Create(model.URLData{
		ShortURL:    types.ShortURL(id),
		OriginalURL: types.RawURL(data.URL),
	}); err != nil {
		var eu *types.ErrUnique
		if errors.As(err, &eu) {
			resp.Header().Set("Content-Type", "application/json")
			// если URL уже существует, то возвращаем короткий URL из базы данных
			resp.WriteHeader(http.StatusConflict)
			shortedURL := s.ResultAddr + "/" + types.ShortURL(eu.ShortURL)
			respData.Result = shortedURL
			if err := json.NewEncoder(resp).Encode(respData); err != nil {
				log.Println("cannot encode response: ", err.Error())
				resp.WriteHeader(http.StatusBadRequest)
				return
			}
			return
		}
		log.Println("cannot write to storage: ", err.Error())
	}
	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(resp).Encode(respData); err != nil {
		log.Println("cannot encode response: ", err.Error())
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	event := &types.Event{
		TimestampEvent: time.Now().UnixNano(),
		Action:         types.ActionCreate,
		UserID:         "",
		URL:            string(data.URL),
	}
	s.subject.Notify(event)
}
