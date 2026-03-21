package shortener

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
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
	content := Hashing(bodyReq)
	outOriginalURL := baseURL + "/" + content
	if err = s.Stor.Create(types.DataURL{
		ShortURL:    types.ShortURL(content),
		OriginalURL: types.RawURL(bodyReq),
	}); err != nil {
		var eu *types.ErrUnique
		if errors.As(err, &eu) {
			// если URL уже существует, то возвращаем короткий URL из базы данных
			resp.WriteHeader(http.StatusConflict)
			shortedURL := s.ResultAddr + "/" + types.ShortURL(eu.ShortURL)
			if _, err = resp.Write([]byte(shortedURL)); err != nil {
				log.Println("cannot write to response: ", err.Error())
				return
			}
		}
		log.Println("cannot write to storage: ", err.Error())
		return
	}
	resp.WriteHeader(http.StatusCreated)
	if _, err = resp.Write([]byte(outOriginalURL)); err != nil {
		s.logger.Error("don't send response because by " + err.Error())
	}
}

// Принимает на вход сокращенный URL,
// возвращает полный URL
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
	location, err := s.Stor.OriginalURL(types.ShortURL(id))
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
		bodyReq []types.BatchRequest
		item    types.BatchRequest
		err     error
		answer  []types.BatchResponse
	)

	reader := bufio.NewReader(req.Body)

	dec := json.NewDecoder(reader)

	if err = dec.Decode(&bodyReq); err != nil {
		s.logger.Error("cannot decode request: " + err.Error())
		resp.WriteHeader(http.StatusBadRequest)
		return
	}
	for _, item = range bodyReq {
		hashURL := Hashing([]byte(item.OriginalURL))
		//сохраняем в хранилище
		if err = s.Stor.Create(types.DataURL{
			ShortURL:    hashURL,
			OriginalURL: item.OriginalURL,
		}); err != nil {
			s.logger.Error("cannot write to storage: " + err.Error())
		}
		shortedURL := s.ResultAddr + "/" + hashURL
		out := types.BatchResponse{
			CorrelationID: item.CorrelationID,
			ShortURL:      shortedURL,
		}
		answer = append(answer, out)
	}

	//устанавливаем тип ответа
	resp.Header().Set("Content-Type", "application/json")
	resp.WriteHeader(http.StatusCreated)
	if err = json.NewEncoder(resp).Encode(answer); err != nil {
		s.logger.Error("cannot encode response: " + err.Error())
		return
	}
}

func Hashing(data []byte) types.ShortURL {
	hashed := sha1.Sum(data)
	shorthed := hashed[:6]
	content := types.ShortURL(hex.EncodeToString(shorthed))
	return content
}
