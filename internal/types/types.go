package types

// Здесь будут храниться пользовательские типы

type RawURL string
type ShortURL string

type DataURL struct {
	ID            uint        `gorm:"not null;primaryKey"`
	ShortURL      ShortURL    `json:"short_url" gorm:"type:varchar(10);not null;uniqueIndex:idx_short_url"`
	OriginalURL   RawURL      `json:"original_url" gorm:"not null;uniqueIndex:idx_original_url"`
	UserProfileID string      `json:"user_profile_uuid" gorm:"type:varchar(36);not null;index:idx_user_profile"`
	UserProfile   UserProfile `json:"user_profile" gorm:"foreignKey:UserProfileID;references:Id;constraint:OnDelete:CASCADE"`
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
