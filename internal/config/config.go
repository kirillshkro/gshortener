package config

import (
	"os"
)

//Конфиг программы. В будущем возможна загрузка из файла или переменных окружения

type Config struct {
	Address    string `json:"address"`
	ShortedURL string `json:"shorted_url"`
}

func NewConfig() *Config {
	var (
		baseAddress string
		shorted     string
		exists      bool
	)

	if baseAddress, exists = os.LookupEnv("SERVER_ADDRESS"); !exists {
		baseAddress = "localhost:8080"
	}
	if shorted, exists = os.LookupEnv("BASE_URL"); !exists {
		shorted = "http://localhost:8080"
	}

	return &Config{
		Address:    baseAddress,
		ShortedURL: shorted,
	}
}
