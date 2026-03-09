package storage

import (
	"sync"

	"github.com/kirillshkro/gshortener/internal/types"
)

//go:generate  mockgen -destination internal/mocks/mock_dbstorage.go -package mocks ./internal/repository/storage IStorage
type Storage struct {
	data map[types.ShortURL]types.RawURL
	mu   sync.Mutex
}

type IStorage interface {
	Data(key types.ShortURL) (types.RawURL, error)
	SetData(reqData types.URLData) error
}

func NewStorage() *Storage {
	return &Storage{
		data: make(map[types.ShortURL]types.RawURL),
	}
}

func (s *Storage) Data(key types.ShortURL) (types.RawURL, error) {
	return s.data[key], nil
}

func (s *Storage) SetData(reqData types.URLData) error {
	s.mu.Lock()
	defer s.mu.Unlock()
	key := reqData.ShortURL
	val := reqData.OriginalURL
	if key != "" && val != "" {
		if _, exist := s.data[key]; !exist {
			s.data[key] = val
		}
	}
	return nil
}
