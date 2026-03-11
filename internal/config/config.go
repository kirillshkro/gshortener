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
		baseAddress     string
		shorted         string
		fileStoragePath string
		dsnDB           string
		cfg             Config
	)

	//Чтение конфига из файла
	if err := cleanenv.ReadConfig(".env", &cfg); err != nil {
		//Чтение конфига из переменных окружения
		if err := cleanenv.ReadEnv(&cfg); err != nil {
			panic(err)
		}
	}

	baseAddress = cfg.Address

	shorted = cfg.ShortedURL

	fileStoragePath = cfg.FileDB

	dsnDB = cfg.DSN

	return &Config{
		Address:    baseAddress,
		ShortedURL: shorted,
		FileDB:     fileStoragePath,
		DSN:        dsnDB,
	}
}
