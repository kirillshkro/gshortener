package config

import (
	"bytes"

	"github.com/kirillshkro/gshortener/internal/handler/shortener"
)

type Config struct {
	Address    string `json:"address"`
	ShortedURL string `json:"shorted_url"`
}

func NewConfig() *Config {
	baseAddress := "localhost:8080"
	shorted := shortener.Hashing(bytes.NewBufferString(baseAddress).Bytes())
	return &Config{
		Address:    baseAddress,
		ShortedURL: shorted,
	}
}
