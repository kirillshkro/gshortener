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

const (
	configFileEnv = "CONFIG"
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
	Address string `env:"ADDRESS" env-default:"localhost:8080" json:"address"`
	// ShortedURL is the base URL for shortened links.
	ShortedURL string `env:"SHORTED_URL" env-default:"http://localhost:8080" json:"short_url"`
	// FileDB is the path to a file where the server should store data.
	FileDB string `env:"FILE_STORAGE_PATH" env-default:"/tmp/shortener.json" json:"file_db_path"`
	// DSN (Data Source Name) is used for database connections.
	DSN string `env:"DATABASE_DSN" json:"database_dsn"`
	// AuditFile is the path to a file where audit logs should be written.
	AuditFile string `env:"AUDIT_FILE" json:"audit_file"`
	// AuditURL is the URL to which audit logs should be sent.
	AuditURL string `env:"AUDIT_URL" json:"audit_url"`
	// EnableHTTPS is flag to enable HTTPS.
	EnableHTTPS bool `env:"ENABLE_HTTPS" env-default:"false" json:"enable_https"`
	// CertFile is the path to the SSL certificate file.
	CertFile string `env:"CERT_FILE" json:"cert_file"`
	// KeyFile is the path to the SSL key file.
	KeyFile string `env:"KEY_FILE" json:"key_file"`
	// ConfigFile is the path to the configuration file.
	ConfigFile string `env:"CONFIG" json:"config_file"`
}

// JSONConfig is a plain struct used for JSON serialization/deserialization.
// All fields are unexported to distinguish it from Config (which has env tags).
type JSONConfig struct {
	Address     string `json:"address"`
	ShortURL    string `json:"short_url"`
	FileDBPath  string `json:"file_db_path"`
	DatabaseDSN string `json:"database_dsn"`
	AuditFile   string `json:"audit_file"`
	AuditURL    string `json:"audit_url"`
	EnableHTTPS bool   `json:"enable_https"`
	CertFile    string `json:"cert_file"`
	KeyFile     string `json:"key_file"`
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

	var jc JSONConfig
	if err := json.Unmarshal(data, &jc); err != nil {
		return fmt.Errorf("failed to parse config file: %w", err)
	}

	// Merge JSON values into Config (JSON has higher priority than env)
	if jc.Address != "" {
		c.Address = jc.Address
	}
	if jc.ShortURL != "" {
		c.ShortedURL = jc.ShortURL
	}
	if jc.FileDBPath != "" {
		c.FileDB = jc.FileDBPath
	}
	if jc.DatabaseDSN != "" {
		c.DSN = jc.DatabaseDSN
	}
	if jc.AuditFile != "" {
		c.AuditFile = jc.AuditFile
	}
	if jc.AuditURL != "" {
		c.AuditURL = jc.AuditURL
	}
	if jc.EnableHTTPS {
		c.EnableHTTPS = jc.EnableHTTPS
	}
	if jc.CertFile != "" {
		c.CertFile = jc.CertFile
	}
	if jc.KeyFile != "" {
		c.KeyFile = jc.KeyFile
	}

	return nil
}

// Save writes the current configuration to a JSON file.
func (c *Config) Save() error {
	if c.ConfigFile == "" {
		return fmt.Errorf("config file path is not set")
	}

	data, err := json.MarshalIndent(JSONConfig{
		Address:     c.Address,
		ShortURL:    c.ShortedURL,
		FileDBPath:  c.FileDB,
		DatabaseDSN: c.DSN,
		AuditFile:   c.AuditFile,
		AuditURL:    c.AuditURL,
		EnableHTTPS: c.EnableHTTPS,
		CertFile:    c.CertFile,
		KeyFile:     c.KeyFile,
	}, "", "  ")
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
