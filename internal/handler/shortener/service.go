package shortener

import (
	"bufio"
	"crypto/sha1"
	"encoding/hex"
	"encoding/json"
	"errors"
	"io"
	"log/slog"
	"net/http"
	"os"
	"strings"
	"time"

	"github.com/kirillshkro/gshortener/internal/config"
	"github.com/kirillshkro/gshortener/internal/config/auth"
	"github.com/kirillshkro/gshortener/internal/handler/shortener/claims"
	"github.com/kirillshkro/gshortener/internal/model"
	"github.com/kirillshkro/gshortener/internal/repository/storage"
	"github.com/kirillshkro/gshortener/internal/service/audit"
	"github.com/kirillshkro/gshortener/internal/types"
)

// Service represents a URL shortening service
type Service struct {
	// ServAddr is the base address of the service
	ServAddr types.RawURL
	// ResultAddr is the address for shortened URLs
	ResultAddr types.ShortURL
	// Stor is the storage interface for URL data
	Stor    storage.IStorage
	logger  *slog.Logger
	subject *audit.Subject
}

// IService defines the interface for all service operations
type IService interface {
	URLEncoder
	URLDecoder
	BatchCreator
	Getter
	Deleter
}

// URLEncoder defines the interface for URL encoding operations
type URLEncoder interface {
	// URLEncode encodes a long URL into a short URL
	URLEncode(resp http.ResponseWriter, req *http.Request)
}

// URLDecoder defines the interface for URL decoding operations
type URLDecoder interface {
	// URLDecode decodes a short URL back to the original URL
	URLDecode(resp http.ResponseWriter, req *http.Request)
}

// BatchCreator defines the interface for batch URL creation operations
type BatchCreator interface {
	// BatchCreateShortURL creates multiple short URLs in batch
	BatchCreateShortURL(resp http.ResponseWriter, req *http.Request)
}

// NewService creates a new service with default address values
// Returns nil if storage initialization fails
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
		subject:    audit.NewSubject(),
	}
}

// NewServiceWithAddr creates a new service with a specified address
// Returns nil if storage initialization fails
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
		subject:    audit.NewSubject(),
	}
}

// NewServiceWithAddrWithAddrShortener creates a new service with specified addresses
// Returns nil if storage initialization fails
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
		subject:    audit.NewSubject(),
	}
}

// URLEncode encodes a long URL into a short URL
// It handles POST requests and returns the short URL in the response body
// Sets appropriate HTTP status codes and headers
func (s Service) URLEncode(resp http.ResponseWriter, req *http.Request) {
	if req.Method != http.MethodPost {
		resp.WriteHeader(http.StatusBadRequest)
		return
	}
	baseURL := s.ResultAddr
	bodyReq, err := io.ReadAll(req.Body)
	if err != nil {
		s.logger.Error("cannot read request: " + err.Error())
		http.Error(resp, "bad request", http.StatusBadRequest)
		return
	}
	resp.Header().Set("Content-Type", "text/plain")
	content := Hashing(bodyReq)
	outOriginalURL := baseURL + "/" + content
	//выдать куку
	var (
		token    string
		userUUID string
		cookie   *http.Cookie
	)
	if cookieExist(req, "auth_cookie") {
		cookie, err := req.Cookie("auth_cookie")
		if err != nil {
			http.Error(resp, "can't get cookie", http.StatusInternalServerError)
			return
		}
		token = cookie.Value
		if token == "" {
			s.refreshUserCookie(resp)
			return
		}
		if userUUID, err = claims.GetUserID(token); err != nil {
			http.Error(resp, "can't get user id", http.StatusInternalServerError)
			return
		}
	} else {
		authCfg := auth.NewAuthConfig()
		authUser := claims.NewAuthUser(authCfg)
		token, err = authUser.Token()
		if err != nil {
			http.Error(resp, "can't get token", http.StatusInternalServerError)
			return
		}
		cookie = &http.Cookie{
			Name:     "auth_cookie",
			Value:    token,
			MaxAge:   3600 * 24 * 7, // 7 дней
			Path:     "/",
			Secure:   false,
			HttpOnly: true,
		}
		http.SetCookie(resp, cookie)
		if userUUID, err = claims.GetUserID(token); err != nil {
			http.Error(resp, "can't get user id", http.StatusInternalServerError)
			return
		}
	}
	//сохраняем в хранилище
	if err = s.Stor.Create(model.URLData{
		ShortURL:    types.ShortURL(content),
		OriginalURL: types.RawURL(bodyReq),
		UserUUID:    userUUID,
	}); err != nil {
		var eu *types.ErrUnique
		if errors.As(err, &eu) {
			// если URL уже существует, то возвращаем короткий URL из базы данных
			resp.WriteHeader(http.StatusConflict)
			s.logger.Info("URL already exists")
			shortedURL := s.ResultAddr + "/" + types.ShortURL(eu.ShortURL)
			if _, err = resp.Write([]byte(shortedURL)); err != nil {
				s.logger.Error("cannot write to response: " + err.Error())
				return
			}
			return
		}
		s.logger.Error("cannot write to storage: " + err.Error())
		return
	}
	resp.WriteHeader(http.StatusCreated)
	if _, err = resp.Write([]byte(outOriginalURL)); err != nil {
		s.logger.Error("don't send response because by " + err.Error())
	}

	event := &types.Event{
		TimestampEvent: time.Now().UnixNano(),
		Action:         types.ActionCreate,
		UserID:         userUUID,
		URL:            string(bodyReq),
	}
	s.subject.Notify(event)
}

// URLDecode decodes a short URL back to the original URL
// It handles GET requests and redirects to the original URL
// Returns appropriate HTTP status codes based on the result
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
		var ad *types.ErrURLDeleted
		if errors.As(err, &ad) {
			http.Error(resp, "URL already deleted", http.StatusGone)
			return
		}
		http.Error(resp, "not found", http.StatusNotFound)
		return
	}

	resp.Header().Set("Location", string(location))
	resp.WriteHeader(http.StatusTemporaryRedirect)

	event := &types.Event{
		TimestampEvent: time.Now().UnixNano(),
		Action:         types.ActionFollow,
		UserID:         "",
		URL:            string(location),
	}
	s.subject.Notify(event)
}

// BatchCreateShortURL creates multiple short URLs in batch
// It expects a JSON array of URL objects in the request body
// Returns a JSON array of created short URLs in the response
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
		if err = s.Stor.Create(model.URLData{
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

// SetSubject sets the audit subject for event tracking
func (s *Service) SetSubject(subj *audit.Subject) {
	if subj != nil {
		s.subject = subj
	}
}

// Hashing generates a short hash of the input data
// Returns a 6-byte hexadecimal string representation of the SHA1 hash
func Hashing(data []byte) types.ShortURL {
	hashed := sha1.Sum(data)
	shorthed := hashed[:6]
	content := types.ShortURL(hex.EncodeToString(shorthed))
	return content
}
