package config

import (
	"flag"
	"os"
)

// Config объект конфига
type Config struct {
	Address         string `env:"SERVER_ADDRESS" envDefault:":8080"`           // адрес запуска HTTP сервера
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"` // базовый URL
	LogLevel        string `env:"LOG_LEVEL" envDefault:"info"`                 // уровень лога
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"./storage"`    // путь до файла хранения сокращенных URL
	DatabaseDSN     string `env:"DATABASE_DSN"`                                // адрес БД
	AuthKey         string `env:"AUTH_KEY" envDefault:"secret_key"`            // секретный ключ для генерации токена
	AuditFile       string `env:"AUDIT_FILE" envDefault:"test.txt"`            // файл хранения аудита
	AuditURL        string `env:"AUDIT_URL" envDefault:""`                     // полный URL удаленного сервера-приёмника
}

// New создание конфигурации
func New() *Config {
	config := &Config{}

	flag.StringVar(&config.Address, "a", ":8080", "Адрес запуска HTTP сервера")
	flag.StringVar(&config.BaseURL, "b", "http://localhost:8080", "Базовый URL")
	flag.StringVar(&config.FileStoragePath, "f", "", "Путь до файла хранения сокращенных URL")
	flag.StringVar(&config.DatabaseDSN, "d", "postgres://username:password@localhost:5432/urls?sslmode=disable", "Адрес БД")
	flag.StringVar(&config.AuditFile, "audit-file", "", "Файл хранения аудита")
	flag.StringVar(&config.AuditURL, "audit-url", "", "Полный URL удаленного сервера-приёмника")

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
