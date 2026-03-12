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
	Address    string `json:"address" env:"SERVER_ADDRESS"`
	ShortedURL string `json:"shorted_url" env:"BASE_URL"`
	FileDB     string `json:"file_db" env:"FILE_STORAGE_PATH"`
	DSN        string `json:"dsn" env:"DATABASE_DSN"`
}

func newConfig() *Config {
	var (
		cfg Config
	)
	//Чтение конфига из переменных окружения
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
