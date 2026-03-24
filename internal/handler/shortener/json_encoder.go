package shortener

import (
	"encoding/json"
	"errors"
	"log"
	"net/http"

	"github.com/kirillshkro/gshortener/internal/types"
)

type JSONEncoder interface {
	CreateShortURL(resp http.ResponseWriter, req *http.Request)
}

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
	if err := s.Stor.Create(types.DataURL{
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
}
