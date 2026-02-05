package storage

import "github.com/kirillshkro/gshortener/internal/types"

type Storage struct {
	data map[types.ShortURL]types.RawURL
}

type IStorage interface {
	Data(key types.ShortURL) types.RawURL
	SetData(key types.ShortURL, val types.RawURL)
}

func NewStorage() *Storage {
	return &Storage{
		data: make(map[types.ShortURL]types.RawURL),
	}
}

func (s Storage) Data(key types.ShortURL) types.RawURL {
	return s.data[key]
}

func (s Storage) SetData(key types.ShortURL, val types.RawURL) {
	if key != "" && val != "" {
		if _, exist := s.data[key]; !exist {
			s.data[key] = val
		}
	}
}
