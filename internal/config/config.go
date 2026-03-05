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
	DSN        string `json:"dsn"`
}

func newConfig() *Config {
	var (
		baseAddress     string
		shorted         string
		fileStoragePath string
		exists          bool
		dsnDb           string
	)

	if baseAddress, exists = os.LookupEnv("SERVER_ADDRESS"); !exists {
		baseAddress = "localhost:8080"
	}

	if shorted, exists = os.LookupEnv("BASE_URL"); !exists {
		shorted = "http://localhost:8080"
	}

	if fileStoragePath, exists = os.LookupEnv("FILE_STORAGE_PATH"); !exists {
		fileStoragePath = "/tmp/shortener.json"
	}

	if dsnDb, exists = os.LookupEnv("DATABASE_DSN"); !exists {
		dsnDb = "postgres://postgres:postgres@localhost:5432/postgres?sslmode=disable"
	}

	return &Config{
		Address:    baseAddress,
		ShortedURL: shorted,
		FileDB:     fileStoragePath,
		DSN:        dsnDb,
	}
}
