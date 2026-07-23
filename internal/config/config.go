// Package config provides configuration for the application. It uses environment variables to configure itself,
// with default values provided if no corresponding environment variable is set.
package config

import (
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

var (
	instance *Config
	once     sync.Once
)

// GetConfig returns the singleton instance of Config. It ensures that only one instance is created,
// even if called concurrently from multiple goroutines.
func GetConfig() *Config {
	once.Do(func() {
		instance = newConfig()
	})
	return instance
}

// Config represents the configuration for the application. It includes fields for various
// types of storage (file, database), server address and URLs, as well as audit logging options.
type Config struct {
	// Address is the address on which the server should listen.
	Address string `env:"ADDRESS" env-default:"localhost:8080"`
	// ShortedURL is the base URL for shortened links.
	ShortedURL string `env:"SHORTED_URL" env-default:"http://localhost:8080"`
	// FileDB is the path to a file where the server should store data.
	FileDB string `env:"FILE_STORAGE_PATH" env-default:"/tmp/shortener.json"`
	// DSN (Data Source Name) is used for database connections.
	DSN string `env:"DATABASE_DSN"`
	// AuditFile is the path to a file where audit logs should be written.
	AuditFile string `env:"AUDIT_FILE"`
	// AuditURL is the URL to which audit logs should be sent.
	AuditURL string `env:"AUDIT_URL"`
}

func newConfig() *Config {
	var cfg Config
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
