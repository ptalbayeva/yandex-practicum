package config

import (
	"flag"
	"os"
)

type UserIDCtx struct{}

type Config struct {
	Address         string `env:"SERVER_ADDRESS" envDefault:":8080"`
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`
	LogLevel        string `env:"LOG_LEVEL" envDefault:"info"`
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"./storage"`
	DatabaseDSN     string `env:"DATABASE_DSN"`
}

func New() *Config {
	config := &Config{}

	flag.StringVar(&config.Address, "a", ":8080", "Адрес запуска HTTP сервера")
	flag.StringVar(&config.BaseURL, "b", "http://localhost:8080", "Базовый URL")
	flag.StringVar(&config.FileStoragePath, "f", "", "Путь до файла хранения сокращенных URL")
	flag.StringVar(&config.DatabaseDSN, "d", "", "Адрес БД")

	flag.Parse()

	if envAddress := os.Getenv("SERVER_ADDRESS"); envAddress != "" {
		config.Address = envAddress
	}

	if envBaseURL := os.Getenv("BASE_URL"); envBaseURL != "" {
		config.BaseURL = envBaseURL
	}

	if envLogLevel := os.Getenv("LOG_LEVEL"); envLogLevel != "" {
		config.LogLevel = envLogLevel
	}

	if envFileStoragePath := os.Getenv("FILE_STORAGE_PATH"); envFileStoragePath != "" {
		config.FileStoragePath = envFileStoragePath
	}

	if config.DatabaseDSN == "" {
		if envDatabaseDSN := os.Getenv("DATABASE_DSN"); envDatabaseDSN != "" {
			config.DatabaseDSN = envDatabaseDSN
		}
	}

	return config
}
