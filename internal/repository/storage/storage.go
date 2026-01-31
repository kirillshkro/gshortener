package storage

type RawURL string
type ShortURL string

type Storage struct {
	data map[ShortURL]RawURL
}

type IStorage interface {
	Data(key ShortURL) RawURL
	SetData(key ShortURL, val RawURL)
}

func NewStorage() *Storage {
	return &Storage{
		data: make(map[ShortURL]RawURL),
	}
}

func (s Storage) Data(key ShortURL) RawURL {
	return s.data[key]
}

func (s Storage) SetData(key ShortURL, val RawURL) {
	if key != "" && val != "" {
		if _, exist := s.data[key]; !exist {
			s.data[key] = val
		}
	}
}
