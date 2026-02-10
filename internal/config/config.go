package config

import "os"

//Конфиг программы. В будущем возможна загрузка из файла или переменных окружения

type Config struct {
	Address    string `json:"address"`
	ShortedURL string `json:"shorted_url"`
}

func NewConfig() *Config {
	var (
		baseAddress string
		shorted     string
	)
	baseAddress = os.Getenv("BASE_ADDRESS")
	shorted = os.Getenv("SHORT_URL")
	if baseAddress == "" {
		baseAddress = "localhost:8080"
	}
	if shorted == "" {
		shorted = "http://localhost:8080"
	}
	return &Config{
		Address:    baseAddress,
		ShortedURL: shorted,
	}
}
