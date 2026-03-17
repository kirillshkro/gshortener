package storage

import (
	"sync"

	"github.com/kirillshkro/gshortener/internal/types"
)

type MemoryStorage struct {
	data map[types.ShortURL]types.RawURL
	mu   sync.Mutex
}

//go:generate mockgen -destination internal/mocks/mock_dbstorage.go -package mocks ./internal/repository/storage IStorage
type IStorage interface {
	Data(key types.ShortURL) (types.RawURL, error)
	SetData(urlData types.URLData) error
	Close() error
	ShortURLGetter
}

type ShortURLGetter interface {
	GetShortURL(key types.RawURL) (types.ShortURL, error)
}

func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[types.ShortURL]types.RawURL),
	}
}

func (s *MemoryStorage) Data(key types.ShortURL) (types.RawURL, error) {
	return s.data[key], nil
}

func (s *MemoryStorage) SetData(urlData types.URLData) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := urlData.ShortURL
	val := urlData.OriginalURL
	if key != "" && val != "" {
		if _, exist := s.data[key]; !exist {
			s.data[key] = val
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
