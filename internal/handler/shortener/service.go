package shortener

import (
	"crypto/sha1"
	"encoding/hex"
	"io"
	"log"
	"net/http"
	"strings"

	"github.com/kirillshkro/gshortener/internal/config"
	"github.com/kirillshkro/gshortener/internal/repository/storage"
	"github.com/kirillshkro/gshortener/internal/types"
)

type Service struct {
	ServAddr   types.RawURL
	ResultAddr types.ShortURL
	Stor       *storage.Storage
	FStor      *storage.FileStorage
}

type IService interface {
	URLEncode(resp http.ResponseWriter, req *http.Request)
	URLDecode(resp http.ResponseWriter, req *http.Request)
}

// Создает сервис со значениями по умолчанию
func NewService() *Service {
	cfg := config.GetConfig()
	stor, err := storage.GetFileStorage(cfg.FileDB)
	if err != nil {
		return nil
	}
	return &Service{
		ServAddr:   types.RawURL("localhost:8080"),
		ResultAddr: types.ShortURL("localhost:8080"),
		Stor:       storage.NewStorage(),
		FStor:      stor,
	}
}

// Создает сервис с заданным IP-адресом и портом
func NewServiceWithAddr(addr types.RawURL) *Service {
	cfg := config.GetConfig()
	stor, err := storage.GetFileStorage(cfg.FileDB)
	if err != nil {
		return nil
	}
	return &Service{
		ServAddr:   addr,
		ResultAddr: types.ShortURL("localhost:8080"),
		Stor:       storage.NewStorage(),
		FStor:      stor,
	}
}

// Создает сервис с заданными IP-адресом и портом, и URL сокращенных ссылок
func NewServiceWithAddrWithAddrShortener(addr types.RawURL, shortAddr types.ShortURL) *Service {
	cfg := config.GetConfig()
	stor, err := storage.GetFileStorage(cfg.FileDB)
	if err != nil {
		return nil
	}
	return &Service{
		ServAddr:   addr,
		ResultAddr: shortAddr,
		Stor:       storage.NewStorage(),
		FStor:      stor,
	}
}

// Принимает на вход URL, возвращает базовый URL сервиса + хэш исходного URL
func (s Service) URLEncode(resp http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}
	baseURL := string(s.ResultAddr)
	bodyReq, err := io.ReadAll(req.Body)
	if err != nil {
		log.Println("cannot read request: ", err.Error())
	}
	resp.Header().Set("Content-Type", "text/plain")
	resp.WriteHeader(http.StatusCreated)
	content := Hashing(bodyReq)
	outData := baseURL + "/" + content
	s.Stor.SetData(types.URLData{
		ShortURL:    types.ShortURL(content),
		OriginalURL: types.RawURL(bodyReq),
	})
	if err := s.FStor.SetData(types.URLData{
		ShortURL:    types.ShortURL(content),
		OriginalURL: types.RawURL(bodyReq),
	}); err != nil {
		http.Error(resp, "unkwown server error: "+err.Error(), http.StatusInternalServerError)
		return
	}
	if _, err = resp.Write([]byte(outData)); err != nil {
		log.Printf("don't send response because by %s\n", err.Error())
	}
}

// Принимает на вход сокращенный URL,
// возвращает исходный
func (s Service) URLDecode(resp http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodGet {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}
	path := req.URL.Path
	id := strings.TrimPrefix(path, "/")
	if id == "" {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}
	loc, _ := s.Stor.Data(types.ShortURL(id))
	location := loc
	resp.Header().Set("Location", string(location))
	resp.WriteHeader(http.StatusTemporaryRedirect)
}

func Hashing(data []byte) string {
	hashed := sha1.Sum(data)
	shorthed := hashed[:6]
	content := hex.EncodeToString(shorthed)
	return content
}
