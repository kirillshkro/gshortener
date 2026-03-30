package config

import (
	"os"
	"sync"
	"time"

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
	CookieConfig
}

type CookieConfig struct {
	SecretKey string        `json:"secret_key" env:"AUTH_SECRET_KEY"`
	Name      string        `json:"name" env:"COOKIE_NAME"`
	MaxAge    time.Duration `json:"max_age" env:"COOKIE_MAX_AGE"`
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
		CookieConfig: CookieConfig{
			SecretKey: os.Getenv("AUTH_SECRET_KEY"),
			Name:      "auth_token",
			MaxAge:    3600 * time.Second,
		},
	}
}
