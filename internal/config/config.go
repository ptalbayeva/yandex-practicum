package config

import (
	"encoding/json"
	"flag"
	"log"
	"os"
)

// Config объект конфига
type Config struct {
	Address         string `env:"SERVER_ADDRESS" envDefault:":8080" json:"address"`    // адрес запуска HTTP сервера
	GRPCAddr        string `env:"GRPC_ADDRESS" envDefault:":3200" json:"grpc_addr"`    // адрес запуска gRPC HTTP сервера
	BaseURL         string `env:"BASE_URL" envDefault:"http://localhost:8080"`         // базовый URL
	LogLevel        string `env:"LOG_LEVEL" envDefault:"info"`                         // уровень лога
	FileStoragePath string `env:"FILE_STORAGE_PATH" envDefault:"./storage"`            // путь до файла хранения сокращенных URL
	DatabaseDSN     string `env:"DATABASE_DSN"`                                        // адрес БД
	AuthKey         string `env:"AUTH_KEY" envDefault:"secret_key"`                    // секретный ключ для генерации токена
	AuditFile       string `env:"AUDIT_FILE" envDefault:"test.txt"`                    // файл хранения аудита
	AuditURL        string `env:"AUDIT_URL" envDefault:""`                             // полный URL удаленного сервера-приёмника
	EnableHttps     bool   `env:"ENABLE_HTTPS" envDefault:"false" json:"enable_https"` // включение HTTPS в веб-сервере
	CertFile        string `env:"CERTFILE" envDefault:"cert.pem" json:"cert_file"`     // сертификат
	KeyFile         string `env:"KEYFILE" envDefault:"key.pem" json:"key_file"`        // ключ
	TrustedSubnet   string `env:"TRUSTED_SUBNET" json:"trusted_subnet"`                // CIDR
}

// New создание конфигурации
func New() *Config {
	config := Config{
		Address:         ":8080",
		GRPCAddr:        ":3200",
		BaseURL:         "http://localhost:8080",
		LogLevel:        "info",
		FileStoragePath: "",
		DatabaseDSN:     "",
		AuthKey:         "secret_key",
		AuditFile:       "",
		AuditURL:        "",
		EnableHttps:     false,
		CertFile:        "cert.pem",
		KeyFile:         "key.pem",
		TrustedSubnet:   "",
	}

	var configPath string

	flag.StringVar(&config.Address, "a", ":8080", "Адрес запуска HTTP сервера")
	flag.StringVar(&config.BaseURL, "b", "http://localhost:8080", "Базовый URL")
	flag.StringVar(&config.FileStoragePath, "f", "", "Путь до файла хранения сокращенных URL")
	flag.StringVar(&config.DatabaseDSN, "d", "", "Адрес БД")
	flag.StringVar(&config.AuditFile, "audit-file", "", "Файл хранения аудита")
	flag.StringVar(&config.AuditURL, "audit-url", "", "Полный URL удаленного сервера-приёмника")
	flag.Bool("s", false, "Включить HTTPS")
	flag.StringVar(&config.CertFile, "cert-file", "", "Файл хранения сертификата")
	flag.StringVar(&config.KeyFile, "key-file", "", "Файл хранения ключа")
	flag.StringVar(&config.TrustedSubnet, "t", "", "Путь к JSON конфигу")
	flag.StringVar(&configPath, "c", os.Getenv("CONFIG"), "Путь к JSON конфигу")

	flag.Parse()

	if configPath != "" {
		file, err := os.Open(configPath)
		if err == nil {
			defer file.Close()
			decoder := json.NewDecoder(file)
			if err = decoder.Decode(config); err != nil {
				log.Printf("error while parsing config file: %v", err)
			}
		}
	}

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

	return &config
}
