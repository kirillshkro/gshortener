// Package types provides the set of data types for managing URLs.
package types

// RawURL is the original URL that is to be shortened. It's a string type, so it can represent any valid URL.
type RawURL string

// ShortURL is a shortened representation of the original URL. It's also a string type, but its value will typically be shorter than the raw URL.
type ShortURL string

// RequestData represents the data sent in a request to create a new short URL. The `json:"url"` tag indicates that this field should be serialized as "url".
type RequestData struct {
	URL RawURL `json:"url"`
}

// ResponseData represents the response from a request to create a new short URL. The `json:"result"` tag indicates that this field should be serialized as "result".
type ResponseData struct {
	Result ShortURL `json:"result"`
}

// BatchRequest represents a batch of requests for creating multiple short URLs. It includes the correlation ID and original URL for each request.
type BatchRequest struct {
	CorrelationID string `json:"correlation_id"`
	OriginalURL   RawURL `json:"original_url"`
}

// BatchResponse represents a batch of responses from creating multiple short URLs. It includes the correlation ID and resulting short URL for each response.
type BatchResponse struct {
	CorrelationID string   `json:"correlation_id"`
	ShortURL      ShortURL `json:"short_url"`
}

// UserURL represents a pair of original and shortened URLs associated with a specific user. It includes the short URL and original URL for each pair.
type UserURL struct {
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}
