package types

// Здесь будут храниться пользовательские типы

type RawURL string
type ShortURL string

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

type UserURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
