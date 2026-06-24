// Package urlgen provides functionality for generating shortened URLs.
package urlgen

import (
	"math/rand"
)

const shortCodeChars = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"

const shortCodeLength = 6

func generateShortCode() string {
	b := make([]byte, shortCodeLength)
	for i := range b {
		b[i] = shortCodeChars[rand.Intn(len(shortCodeChars))]
	}
	return string(b)
}

// GenerateURL generates a valid random shortened URL.
// baseURL is the base of the shortening service, e.g. "https://short.url".
// The function appends a randomly generated short code to the base URL.
func GenerateURL(baseURL string) string {
	return baseURL + "/" + generateShortCode()
}
