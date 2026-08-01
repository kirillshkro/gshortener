// Package config provides configuration for the application. It uses environment variables to configure itself,
// with default values provided if no corresponding environment variable is set.
package config

import (
	"encoding/json"
	"fmt"
	"os"
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
	DSN string `env:"DATABASE_DSN" json:"database_dsn"`
	// AuditFile is the path to a file where audit logs should be written.
	AuditFile string `env:"AUDIT_FILE" json:"audit_file"`
	// AuditURL is the URL to which audit logs should be sent.
	AuditURL string `env:"AUDIT_URL" json:"audit_url"`
	// EnableHTTPS is flag to enable HTTPS.
	EnableHTTPS bool `env:"ENABLE_HTTPS" env-default:"false"`
	// CertFile is the path to the SSL certificate file.
	CertFile string `env:"CERT_FILE"`
	// KeyFile is the path to the SSL key file.
	KeyFile string `env:"KEY_FILE"`
	// ConfigFile is the path to the configuration file.
	ConfigFile string `env:"CONFIG"`
	// GRPCAddress is the address on which the gRPC server should listen.
	GRPCAddress string `env:"GRPC_ADDRESS" env-default:""`
}

// newConfig reads configuration from environment variables and, if a config file
// path is provided via CONFIG env, merges values from that JSON file.
func newConfig() *Config {
	var cfg Config
	if err := cleanenv.ReadEnv(&cfg); err != nil {
		panic(err)
	}

	// Read from JSON config file if path is set
	if cfg.ConfigFile != "" {
		if err := cfg.ReadFromJSON(cfg.ConfigFile); err != nil {
			panic(fmt.Errorf("failed to read config file %q: %w", cfg.ConfigFile, err))
		}
	}

	return &cfg
}

// ReadFromJSON reads configuration from a JSON file and merges it into the Config.
// JSON values override environment values (higher priority).
func (c *Config) ReadFromJSON(filePath string) error {
	data, err := os.ReadFile(filePath)
	if err != nil {
		return fmt.Errorf("failed to read config file: %w", err)
	}

	var tmpConfig Config
	if err := json.Unmarshal(data, &tmpConfig); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	c = &tmpConfig
	return nil
}

// Save writes the current configuration to a JSON file.
func (c *Config) Save() error {
	if c.ConfigFile == "" {
		return fmt.Errorf("config file path is not set")
	}

	data, err := json.MarshalIndent(c, "", "  ")
	if err != nil {
		return fmt.Errorf("failed to marshal config: %w", err)
	}

	f, err := os.OpenFile(c.ConfigFile, os.O_RDWR|os.O_CREATE|os.O_TRUNC, 0644)
	if err != nil {
		return fmt.Errorf("failed to open config file: %w", err)
	}
	if _, err := f.Write(data); err != nil {
		f.Close()
		return fmt.Errorf("failed to write config file: %w", err)
	}
	return f.Close()
}
