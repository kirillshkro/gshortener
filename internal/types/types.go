package types

// Здесь будут храниться пользовательские типы

type RawURL string
type ShortURL string

type DataURL struct {
	ID          uint        `gorm:"not null;primaryKey"`
	ShortURL    ShortURL    `json:"short_url" gorm:"type:varchar(10);not null;uniqueIndex:idx_short_url"`
	OriginalURL RawURL      `json:"original_url" gorm:"not null;uniqueIndex:idx_original_url"`
	UserProfile UserProfile `json:"user_profile" gorm:"foreignKey"`
}

type FileData struct {
	UUID string `json:"uuid"`
	DataURL
}

type RequestData struct {
	URL RawURL `json:"url"`
}

type ResponseData struct {
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
