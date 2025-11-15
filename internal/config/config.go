package config

import (
	"flag"
	"os"
)

type Config struct {
	Address string `env:"SERVER_ADDRESS" envDefault:":8080"`
	BaseURL string `env:"BASE_URL" envDefault:"http://localhost:8080"`
}

func New() *Config {
	config := &Config{}

	flag.StringVar(&config.Address, "a", ":8080", "Адрес запуска HTTP сервера")
	flag.StringVar(&config.BaseURL, "b", "http://localhost:8080", "Базовый URL")

	flag.Parse()

	if envAddress := os.Getenv("SERVER_ADDRESS"); envAddress != "" {
		config.Address = envAddress
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		config.BaseURL = envBaseURL
	}

	return config
}
