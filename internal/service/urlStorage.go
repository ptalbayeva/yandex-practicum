package service

import (
	"encoding/json"
	"os"

	"github.com/google/uuid"
)

type StorageService struct {
	fileStoragePath string
}

type Event struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewStorageService(fileStoragePath string) *StorageService {
	return &StorageService{fileStoragePath: fileStoragePath}
}

func newEvent(originalURL string, shortURL string) *Event {
	return &Event{
		UUID:        uuid.NewString(),
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}
}

func (s *StorageService) Save(originalURL string, shortURL string) error {
	event := newEvent(originalURL, shortURL)
	data, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		return err
	}

	file, err := os.OpenFile(s.fileStoragePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
	if err != nil {
		return err
	}

	defer file.Close()

	_, err = file.Write(data)

	if err != nil {
		return err
	}

	return nil
}
