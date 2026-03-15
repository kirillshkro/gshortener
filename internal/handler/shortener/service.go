package shortener

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"io"
	"log"
	"log/slog"
	"net/http"
	"os"
	"strings"

	"github.com/kirillshkro/gshortener/internal/config"
	"github.com/kirillshkro/gshortener/internal/repository/storage"
	"github.com/kirillshkro/gshortener/internal/types"
)

type Service struct {
	ServAddr   types.RawURL
	ResultAddr types.ShortURL
	Stor       storage.IStorage
	logger     *slog.Logger
}

type IService interface {
	URLEncoder
	URLDecoder
	BatchCreator
}

type URLEncoder interface {
	URLEncode(resp http.ResponseWriter, req *http.Request)
}

type URLDecoder interface {
	URLDecode(resp http.ResponseWriter, req *http.Request)
}

type BatchCreator interface {
	BatchCreateShortURL(resp http.ResponseWriter, req *http.Request)
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
		Stor:       stor,
		logger:     slog.New(slog.NewTextHandler(os.Stderr, nil)),
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
		Stor:       stor,
		logger:     slog.New(slog.NewTextHandler(os.Stderr, nil)),
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
		Stor:       stor,
		logger:     slog.New(slog.NewTextHandler(os.Stderr, nil)),
	}
}

// Принимает на вход URL, возвращает базовый URL сервиса + хэш исходного URL
func (s Service) URLEncode(resp http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}
	baseURL := s.ResultAddr
	bodyReq, err := io.ReadAll(req.Body)
	if err != nil {
		log.Println("cannot read request: ", err.Error())
	}
	resp.Header().Set("Content-Type", "text/plain")
	resp.WriteHeader(http.StatusCreated)
	content := Hashing(bodyReq)
	outData := baseURL + "/" + content
	if err = s.Stor.SetData(types.URLData{
		ShortURL:    types.ShortURL(content),
		OriginalURL: types.RawURL(bodyReq),
	}); err != nil {
		log.Println("cannot write to storage: ", err.Error())
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
	location, err := s.Stor.Data(types.ShortURL(id))
	if err != nil {
		http.Error(resp, "not found", http.StatusNotFound)
		return
	}
	resp.Header().Set("Location", string(location))
	resp.WriteHeader(http.StatusTemporaryRedirect)
}

func (s Service) BatchCreateShortURL(resp http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}

	var (
		item   types.BatchRequest
		err    error
		answer []types.BatchResponse
	)

	reader := bufio.NewReader(req.Body)

	dec := json.NewDecoder(reader)

	for {
		if err = dec.Decode(&item); err != nil {
			if err == io.EOF {
				break
			}
			s.logger.Error("cannot decode request: ", err.Error())
			resp.WriteHeader(http.StatusBadRequest)
			return
		}
		hashURL := Hashing([]byte(item.OriginalURL))
		shortedURL := s.ResultAddr + "/" + hashURL
		out := types.BatchResponse{
			CorrelationID: item.CorrelationID,
			ShortURL:      shortedURL,
		}
		answer = append(answer, out)
	}
	resp.Header().Set("Content-Type", "application/json")
	if err = json.NewEncoder(resp).Encode(answer); err != nil {
		s.logger.Error("cannot encode response: ", err.Error())
		return
	}
	resp.WriteHeader(http.StatusCreated)
}

func Hashing(data []byte) types.ShortURL {
	hashed := sha1.Sum(data)
	shorthed := hashed[:6]
	content := types.ShortURL(hex.EncodeToString(shorthed))
	return content
}
