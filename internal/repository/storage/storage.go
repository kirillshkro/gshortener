package storage

import (
	"fmt"
	"sync"

	"github.com/kirillshkro/gshortener/internal/types"
	"github.com/kirillshkro/gshortener/internal/types/model"
)

type MemoryStorage struct {
	data map[types.ShortURL]types.RawURL
	mu   sync.Mutex
}

//go:generate mockgen -destination internal/mocks/mock_dbstorage.go -package mocks ./internal/repository/storage IStorage
type IStorage interface {
	OriginalURL(key types.ShortURL) (types.RawURL, error)
	Create(urlOriginalURL model.URLData) error
	Close() error
	UserGetter
	Deleter
}

type Deleter interface {
	DeleteUserURL(userID string, url types.ShortURL) error
}

type UserGetter interface {
	GetUserURLs(userUUID string) ([]types.UserURL, error)
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[types.ShortURL]types.RawURL),
	}
}

func (s *MemoryStorage) OriginalURL(key types.ShortURL) (types.RawURL, error) {
	return s.data[key], nil
}

func (s *MemoryStorage) Create(req model.URLData) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := req.ShortURL
	val := req.OriginalURL
	if key != "" && val != "" {
		if _, exist := s.data[key]; !exist {
			s.data[key] = val
		} else {
			return &types.ErrUnique{
				CauseURL: val,
				ShortURL: key,
				Err:      fmt.Errorf("error duplicate value %s", val),
			}
		}
	}
	return nil
}

func (s *MemoryStorage) Close() error {
	return nil
}

func (s *MemoryStorage) GetShortURL(key types.RawURL) (types.ShortURL, error) {
	for k, v := range s.data {
		if v == key {
			return k, nil
		}
	}
	return "", types.ErrNotFound
}

func (s *MemoryStorage) GetUserURLs(userUUID string) ([]types.UserURL, error) {
	return []types.UserURL{}, nil
}

func (s *MemoryStorage) DeleteUserURLs(userID string, shortURL types.ShortURL) error {
	return nil
}
