package config

import (
	"os"
	"sync"
)

var (
	instance *Config
	once     sync.Once
)

func GetConfig() *Config {
	once.Do(func() {
		instance = newConfig()
	})
	return instance
}

//Конфиг программы.

type Config struct {
	Address    string `json:"address"`
	ShortedURL string `json:"shorted_url"`
	FileDB     string `json:"file_db"`
}

func newConfig() *Config {
	var (
		baseAddress     string
		shorted         string
		fileStoragePath string
		exists          bool
	)

	if baseAddress, exists = os.LookupEnv("SERVER_ADDRESS"); !exists {
		baseAddress = "localhost:8080"
	}

	if shorted, exists = os.LookupEnv("BASE_URL"); !exists {
		shorted = "http://localhost:8080"
	}

	if fileStoragePath, exists = os.LookupEnv("FILE_STORAGE_PATH"); !exists {
		fileStoragePath = "/var/run/shortener.json"
	}

	return &Config{
		Address:    baseAddress,
		ShortedURL: shorted,
		FileDB:     fileStoragePath,
	}
}
