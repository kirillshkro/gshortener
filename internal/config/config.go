package config

//Конфиг программы
//В будущем возможна загрузка из файла или переменных окружения

type Config struct {
	Address    string `json:"address"`
	ShortedURL string `json:"shorted_url"`
}

func NewConfig() *Config {
	baseAddress := "localhost:8080"
	shorted := "http://localhost:8080"
	return &Config{
		Address:    baseAddress,
		ShortedURL: shorted,
	}
}
