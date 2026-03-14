package config

import (
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
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
	Address    string `json:"address" env:"ADDRESS" env-default:"localhost:8080"`
	ShortedURL string `json:"shorted_url" env:"SHORTED_URL" env-default:"http://localhost:8080"`
	FileDB     string `json:"file_db" env:"FILE_STORAGE_PATH" env-default:"/tmp/shortener.json"`
	DSN        string `json:"dsn" env:"DATABASE_DSN"`
}

func newConfig() *Config {
	var (
		cfg Config
	)

	if err := cleanenv.ReadEnv(&cfg); err != nil {
		panic(err)
	}

	return &Config{
		Address:    cfg.Address,
		ShortedURL: cfg.ShortedURL,
		FileDB:     cfg.FileDB,
		DSN:        cfg.DSN,
	}
}
