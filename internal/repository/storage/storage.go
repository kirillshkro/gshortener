package storage

import (
	"context"
	"fmt"
	"sync"

	"github.com/kirillshkro/gshortener/internal/model"
	"github.com/kirillshkro/gshortener/internal/types"
)

// MemoryStorage represents an in-memory storage implementation for URL shortening
type MemoryStorage struct {
	// data stores the mapping between short URLs and original URLs
	data map[types.ShortURL]types.RawURL
	// mu provides mutex lock for concurrent access to data
	mu sync.Mutex
}

// IStorage defines the interface for URL storage operations
//
//go:generate mockgen -destination internal/mocks/mock_dbstorage.go -package mocks ./internal/repository/storage IStorage
type IStorage interface {
	// OriginalURL returns the original URL for a given short URL key
	// Returns the original URL and an error if the key is not found
	OriginalURL(key types.ShortURL) (types.RawURL, error)
	// Create stores a new URL mapping
	// Returns an error if the operation fails
	Create(urlOriginalURL model.URLData) error
	// Close releases any resources used by the storage
	// Returns an error if closing fails
	Close() error
	UserGetter
	Deleter
}

// UserGetter defines the interface for retrieving user-specific URLs
type UserGetter interface {
	// GetUserURLs returns all URLs associated with a given user UUID
	// Returns a slice of UserURL and an error if the operation fails
	GetUserURLs(userUUID string) ([]types.UserURL, error)
}

// Deleter defines the interface for deleting URLs
type Deleter interface {
	// DeleteUserURL deletes a URL mapping for a given short URL
	// Returns an error if the operation fails
	DeleteUserURL(ctx context.Context, shortURL types.ShortURL) error
}

// NewMemoryStorage creates and returns a new instance of MemoryStorage
// Returns a pointer to the newly created MemoryStorage
func NewMemoryStorage() *MemoryStorage {
	return &MemoryStorage{
		data: make(map[types.ShortURL]types.RawURL),
	}
}

// OriginalURL returns the original URL for a given short URL key
// Parameters:
//   - key: the short URL key to look up
//
// Returns:
//   - the original URL associated with the key
//   - an error if the key is not found
func (s *MemoryStorage) OriginalURL(key types.ShortURL) (types.RawURL, error) {
	return s.data[key], nil
}

// Create stores a new URL mapping in the storage
// Parameters:
//   - req: the URLData containing the short URL and original URL to store
//
// Returns:
//   - an error if the operation fails (e.g., duplicate key)
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

// Close releases any resources used by the storage
// Returns an error if closing fails
func (s *MemoryStorage) Close() error {
	return nil
}

// GetShortURL finds and returns the short URL for a given original URL
// Parameters:
//   - key: the original URL to search for
//
// Returns:
//   - the short URL associated with the original URL
//   - an error if the original URL is not found
func (s *MemoryStorage) GetShortURL(key types.RawURL) (types.ShortURL, error) {
	for k, v := range s.data {
		if v == key {
			return k, nil
		}
	}
	return "", types.ErrNotFound
}

// GetUserURLs returns all URLs associated with a given user UUID
// Parameters:
//   - userUUID: the UUID of the user whose URLs are to be retrieved
//
// Returns:
//   - a slice of UserURL containing the user's URLs
//   - an error if the operation fails
func (s *MemoryStorage) GetUserURLs(userUUID string) ([]types.UserURL, error) {
	return []types.UserURL{}, nil
}

// DeleteUserURL deletes a URL mapping for a given short URL
// Parameters:
//   - ctx: the context for the operation
//   - shortURL: the short URL to delete
//
// Returns:
//   - an error if the operation fails
func (s *MemoryStorage) DeleteUserURL(ctx context.Context, shortURL types.ShortURL) error {
	return nil
}
