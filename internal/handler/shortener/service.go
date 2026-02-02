package shortener

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"log"
	"net/http"

	"github.com/gorilla/mux"
	"github.com/kirillshkro/gshortener/internal/repository/storage"
)

type Service struct {
	ServAddr   storage.RawURL
	ResultAddr storage.ShortURL
	Stor       *storage.Storage
}

type IService interface {
	URLEncode(resp http.ResponseWriter, req *http.Request)
	URLDecode(resp http.ResponseWriter, req *http.Request)
}

func NewService() *Service {
	return &Service{
		ServAddr:   storage.RawURL("localhost:8080"),
		ResultAddr: storage.ShortURL("localhost:8000"),
		Stor:       storage.NewStorage(),
	}
}

func NewServiceWithAddr(addr storage.RawURL) *Service {
	return &Service{
		ServAddr:   addr,
		ResultAddr: storage.ShortURL("localhost:8000"),
		Stor:       storage.NewStorage(),
	}
}

func NewServiceWithAddrWithAddrShortener(addr storage.RawURL, shortAddr storage.ShortURL) *Service {
	return &Service{
		ServAddr:   addr,
		ResultAddr: shortAddr,
		Stor:       storage.NewStorage(),
	}
}

// Принимает на вход URL, возвращает базовый URL сервиса + хэш исходного URL
func (s Service) URLEncode(resp http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}
	defer req.Body.Close()
	bodyReq, err := io.ReadAll(req.Body)
	if err != nil {
		log.Println("cannot read request: ", err.Error())
	}
	resp.Header().Set("Content-Type", "text/plain")
	resp.WriteHeader(http.StatusCreated)
	content := Hashing(bodyReq)
	outData := string(s.ResultAddr) + "/" + content
	s.Stor.SetData(storage.ShortURL(content), storage.RawURL(bodyReq))
	if _, err = resp.Write([]byte("https://" + outData)); err != nil {
		log.Printf("don't send response because by %s\n", err.Error())
	}
}

func (s Service) URLDecode(resp http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}
	vars := mux.Vars(req)
	id := vars["id"]
	pattern := id
	original := s.Stor.Data(storage.ShortURL(pattern))
	resp.Header().Set("Location", string(original))
	resp.WriteHeader(http.StatusTemporaryRedirect)
}

func Hashing(data []byte) string {
	hashed := sha1.Sum(data)
	shorthed := hashed[:6]
	content := hex.EncodeToString(shorthed)
	return content
}
