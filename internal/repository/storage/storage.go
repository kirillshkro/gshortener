package storage

import (
	"context"
	"fmt"
	"sync"

	"github.com/kirillshkro/gshortener/internal/model"
	"github.com/kirillshkro/gshortener/internal/types"
)

type MemoryStorage struct {
	data   map[types.ShortURL]types.RawURL
	userID map[types.ShortURL]string
	mu     sync.Mutex
}

//go:generate mockgen -destination internal/mocks/mock_dbstorage.go -package mocks ./internal/repository/storage IStorage
type IStorage interface {
	// возвращаем исходный URL и ошибку
	OriginalURL(key types.ShortURL) (types.RawURL, error)
	Create(urlOriginalURL model.URLData) error
	Close() error
	UserGetter
	Deleter
}

type UserGetter interface {
	GetUserURLs(userUUID string) ([]types.UserURL, error)
}

type Deleter interface {
	DeleteUserURL(ctx context.Context, shortURL types.ShortURL) error
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data:   make(map[types.ShortURL]types.RawURL),
		userID: make(map[types.ShortURL]string),
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
			s.userID[key] = req.UserUUID
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
	s.mu.Lock()
	defer s.mu.Unlock()

	var result []types.UserURL
	for shortURL, originalURL := range s.data {
		if s.userID[shortURL] == userUUID {
			result = append(result, types.UserURL{
				ShortURL:    string(shortURL),
				OriginalURL: string(originalURL),
			})
		}
	}
	return result, nil
}

func (s *MemoryStorage) DeleteUserURL(ctx context.Context, shortURL types.ShortURL) error {
	return nil
}
