package types

// Здесь будут храниться пользовательские типы

type RawURL string
type ShortURL string

type FileData struct {
	UUID        uint     `json:"uuid"`
	ShortURL    ShortURL `json:"short_url"`
	OriginalURL RawURL   `json:"original_url"`
}

type TStor []FileData
