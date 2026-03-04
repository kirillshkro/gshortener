package shortener

import (
	"encoding/json"
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
	respData.Result = string(s.ResultAddr) + "/" + id
	s.Stor.SetData(types.ShortURL(id), types.RawURL(data.URL))
	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(http.StatusCreated)
	if err := json.NewEncoder(resp).Encode(respData); err != nil {
		log.Println("cannot encode response: ", err.Error())
		resp.WriteHeader(http.StatusBadRequest)
		return
	}
}
