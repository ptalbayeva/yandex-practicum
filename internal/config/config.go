package config

import (
	"flag"
)

type Config struct {
	Address string
	BaseURL string
}

func New() *Config {
	config := &Config{}

	flag.StringVar(&config.Address, "a", ":8081", "Адрес запуска HTTP сервера")
	flag.StringVar(&config.BaseURL, "b", "http://localhost:8081", "Базовый URL")

	flag.Parse()

	return config
}
