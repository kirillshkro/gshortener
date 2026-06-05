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
	Address    string `env:"ADDRESS" env-default:"localhost:8080"`
	ShortedURL string `env:"SHORTED_URL" env-default:"http://localhost:8080"`
	FileDB     string `env:"FILE_STORAGE_PATH" env-default:"/tmp/shortener.json"`
	DSN        string `env:"DATABASE_DSN"`
	AuditFile  string `env:"AUDIT_FILE"`
	AuditURL   string `env:"AUDIT_URL"`
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
		AuditFile:  cfg.AuditFile,
		AuditURL:   cfg.AuditURL,
	}
}
