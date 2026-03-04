package types

// Здесь будут храниться пользовательские типы

type RawURL string
type ShortURL string

type FileData struct {
	UUID        int64    `json:"uuid"`
	ShortURL    ShortURL `json:"short_url"`
	OriginalURL RawURL   `json:"original_url"`
}

type TStor []FileData

type RequestData struct {
	URL string `json:"url"`
}

type ResponseData struct {
	Result string `json:"result"`
}
