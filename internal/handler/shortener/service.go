package shortener

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/kirillshkro/gshortener/internal/repository/storage"
	"github.com/kirillshkro/gshortener/internal/types"
)

type Service struct {
	ServAddr   types.RawURL
	ResultAddr types.ShortURL
	Stor       *storage.Storage
}

type IService interface {
	URLEncode(resp http.ResponseWriter, req *http.Request)
	URLDecode(resp http.ResponseWriter, req *http.Request)
}

func NewService() *Service {
	return &Service{
		ServAddr:   types.RawURL("localhost:8080"),
		ResultAddr: types.ShortURL("localhost:8000"),
		Stor:       storage.NewStorage(),
	}
}

func NewServiceWithAddr(addr types.RawURL) *Service {
	return &Service{
		ServAddr:   addr,
		ResultAddr: types.ShortURL("localhost:8000"),
		Stor:       storage.NewStorage(),
	}
}

func NewServiceWithAddrWithAddrShortener(addr types.RawURL, shortAddr types.ShortURL) *Service {
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
	baseURL := string(s.ResultAddr)
	defer req.Body.Close()
	bodyReq, err := io.ReadAll(req.Body)
	if err != nil {
		log.Println("cannot read request: ", err.Error())
	}
	resp.Header().Set("Content-Type", "text/plain")
	resp.WriteHeader(http.StatusCreated)
	content := Hashing(bodyReq)
	outData := baseURL + "/" + content
	s.Stor.SetData(types.ShortURL(content), types.RawURL(bodyReq))
	if _, err = resp.Write([]byte(outData)); err != nil {
		log.Printf("don't send response because by %s\n", err.Error())
	}
}

func (s Service) URLDecode(resp http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}
	path := req.URL.Path
	id := strings.TrimPrefix(path, "/")
	location := s.Stor.Data(types.ShortURL(id))
	resp.Header().Set("Location", string(location))
	resp.WriteHeader(http.StatusTemporaryRedirect)
}

func Hashing(data []byte) string {
	hashed := sha1.Sum(data)
	shorthed := hashed[:6]
	content := hex.EncodeToString(shorthed)
	return content
}
