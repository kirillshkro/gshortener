package types

// Здесь будут храниться пользовательские типы

type RawURL string
type ShortURL string

type URLData struct {
	ShortURL    ShortURL `json:"short_url"`
	OriginalURL RawURL   `json:"original_url"`
}

type FileData struct {
	UUID string `json:"uuid"`
	URLData
}

type RequestData struct {
	URL string `json:"url"`
}

type ResponseData struct {
	Result string `json:"result"`
}

type BatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   string `json:"original_url"`
}

type BatchResponse struct {
	CorrelationID string `json:"correlation_id"`
	ShortURL      string `json:"short_url"`
}
