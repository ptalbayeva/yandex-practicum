package config

import (
	"flag"

	"github.com/ilyakaznacheev/cleanenv"
)

type Config struct {
	Server struct {
		Address string `yaml:"address" env:"SERVER_ADDRESS" env-default:":8080"`
		BaseURL string `yaml:"baseUrl" env:"BASE_URL" env-default:"http://localhost:8080"`
	} `yaml:"server"`
	Database struct {
		DSN string `yaml:"dsn" env:"DATABASE_DSN"`
	} `yaml:"database"`
	FileStoragePath string `yaml:"fileStoragePath" env:"FILE_STORAGE_PATH" env-default:"./storage"`
	LogLevel        string `yaml:"logLevel" env:"LOG_LEVEL" env-default:"info"`
}

func New(path string) (*Config, error) {
	config := &Config{}

	if err := cleanenv.ReadConfig(path, config); err != nil {
		return nil, err
	}

	flag.StringVar(&config.Server.Address, "a", ":8080", "Адрес запуска HTTP сервера")
	flag.StringVar(&config.Server.BaseURL, "b", "http://localhost:8080", "Базовый URL")
	flag.StringVar(&config.FileStoragePath, "f", "./storage", "Путь до файла хранения сокращенных URL")
	flag.StringVar(&config.Database.DSN, "d", "", "Адрес БД")

	flag.Parse()

	if err := cleanenv.ReadEnv(config); err != nil {
		return nil, err
	}

	return config, nil
}
