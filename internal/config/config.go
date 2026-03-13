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
	Address    string `json:"address"`
	ShortedURL string `json:"shorted_url"`
	FileDB     string `json:"file_db"`
	DSN        string `json:"dsn"`
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
