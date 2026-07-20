// Package urlgen предоставляет функции для генерации случайных коротких URL-адресов.
// Используется для создания уникальных идентификаторов при сокращении URL.
package urlgen

import (
	"math/rand"
)

const (
	shortCodeChars  = "abcdefghijklmnopqrstuvwxyzABCDEFGHIJKLMNOPQRSTUVWXYZ0123456789"
	shortCodeLength = 6
)

func generateShortCode() string {
	b := make([]byte, shortCodeLength)
	for i := range b {
		b[i] = shortCodeChars[rand.Intn(len(shortCodeChars))]
	}
	return string(b)
}

// GenerateURL генерирует случайный короткий URL на основе базового адреса.
// Формирует полный URL путем объединения базового адреса и сгенерированного короткого кода.
//
// Пример использования:
//
//	url := GenerateURL("https://short.url")
//	// Результат: "https://short.url/abc123"
//
// Параметры:
//   - baseURL: базовый адрес сервиса сокращения URL (например, "https://short.url")
//
// Возвращает:
//   - string: полный короткий URL в формате baseURL + "/" + shortCode
func GenerateURL(baseURL string) string {
	return baseURL + "/" + generateShortCode()
}
