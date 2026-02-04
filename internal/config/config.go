package config

type Config struct {
	Address    string `json:"address"`
	ShortedURL string `json:"shorted_url"`
}

func NewConfig() *Config {
	baseAddress := "localhost:8080"
	shorted := "http://localhost:8888"
	return &Config{
		Address:    baseAddress,
		ShortedURL: shorted,
	}
}
