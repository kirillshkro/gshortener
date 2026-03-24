package urlgen

import (
	"math/rand"
)

const (
	shortCodeChars  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	shortCodeLength = 6
)

// generateShortCode generates a random short code of fixed length
func generateShortCode() string {
	b := make([]byte, shortCodeLength)
	for i := range b {
		b[i] = shortCodeChars[rand.Intn(len(shortCodeChars))]
	}
	return string(b)
}

// GenerateURL generates a valid random shortened URL
// baseURL is the base of the shortening service, e.g. "https://short.url"
func GenerateURL(baseURL string) string {
	return baseURL + "/" + generateShortCode()
}
