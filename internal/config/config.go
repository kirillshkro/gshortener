// Package config provides configuration management for the application.
package config

import (
	"sync"

	"github.com/ilyakaznacheev/cleanenv"
)

// instance holds the singleton instance of Config.
var (
	instance *Config
	once     sync.Once
)

// GetConfig returns the singleton instance of Config.
// It ensures that only one instance of Config is created during the application lifetime.
func GetConfig() *Config {
	once.Do(func() {
		instance = newConfig()
	})
	return instance
}

// Config represents the application configuration.
// It contains all the necessary settings for the application to run properly.
type Config struct {
	// Address is the host and port on which the application will listen.
	// Default value is "localhost:8080".
	Address string `env:"ADDRESS" env-default:"localhost:8080"`

	// ShortedURL is the base URL for shortened URLs.
	// Default value is "http://localhost:8080".
	ShortedURL string `env:"SHORTED_URL" env-default:"http://localhost:8080"`

	// FileDB is the path to the file storage for the application.
	// Default value is "/tmp/shortener.json".
	FileDB string `env:"FILE_STORAGE_PATH" env-default:"/tmp/shortener.json"`

	// DSN is the Data Source Name for the database connection.
	// This field is required and should contain the database connection string.
	DSN string `env:"DATABASE_DSN"`

	// AuditFile is the path to the audit log file.
	// This field is optional and can be left empty if audit logging to file is not needed.
	AuditFile string `env:"AUDIT_FILE"`

	// AuditURL is the URL for audit logging.
	// This field is optional and can be left empty if audit logging to URL is not needed.
	AuditURL string `env:"AUDIT_URL"`
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
