package repository

import (
	"bufio"
	"encoding/json"
	"errors"
	"os"

	"github.com/google/uuid"
	"github.com/yandex-practicum/shorten-url/internal/model"
)

type FileRepository struct {
	fileStoragePath string
}

type Event struct {
	UUID        string `json:"uuid"`
	ShortURL    string `json:"short_url"`
	OriginalURL string `json:"original_url"`
}

func NewFileRepository(fileStoragePath string) *FileRepository {
	return &FileRepository{fileStoragePath: fileStoragePath}
}

func newEvent(originalURL string, shortURL string) *Event {
	return &Event{
		UUID:        uuid.NewString(),
		ShortURL:    shortURL,
		OriginalURL: originalURL,
	}
}

func (f *FileRepository) Save(u *model.URL) error {
	event := newEvent(u.Original, u.Code)
	data, err := json.MarshalIndent(event, "", "  ")
	if err != nil {
		return err
	}

	file, err := os.OpenFile(f.fileStoragePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0777)
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

func (f *FileRepository) FindByCode(code string) (*model.URL, error) {
	file, err := os.Open(f.fileStoragePath)

	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	for scanner.Scan() {
		var event Event

		data := scanner.Bytes()
		fail := json.Unmarshal(data, &event)
		if fail != nil {
			continue
		}

		if event.ShortURL == code {
			return model.NewURL(event.OriginalURL, event.ShortURL), nil
		}
	}

	if fail := scanner.Err(); fail != nil {
		return nil, fail
	}

	return nil, errors.New("not found")
}
