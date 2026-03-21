package types

// Здесь будут храниться пользовательские типы

type RawURL string
type ShortURL string

type DataURL struct {
	ID          uint     `gorm:"not null;primaryKey"`
	ShortURL    ShortURL `json:"short_url" gorm:"not null;uniqueIndex"`
	OriginalURL RawURL   `json:"original_url" gorm:"not null;uniqueIndex"`
}

type FileOriginalURL struct {
	UUID string `json:"uuid"`
	DataURL
}

type RequestOriginalURL struct {
	URL RawURL `json:"url"`
}

type ResponseOriginalURL struct {
	Result ShortURL `json:"result"`
}

type BatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   RawURL `json:"original_url"`
}

type BatchResponse struct {
	CorrelationID string   `json:"correlation_id"`
	ShortURL      ShortURL `json:"short_url"`
}
