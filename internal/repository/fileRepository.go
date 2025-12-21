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
	UserID      string `json:"user_id,omitempty"`
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
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	file, err := os.OpenFile(f.fileStoragePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	defer file.Close()

	if _, fail := file.Write(append(data, '\n')); fail != nil {
		return fail
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
			return nil, fail
		}

		if event.ShortURL == code {
			return model.NewURL(event.ShortURL, event.OriginalURL, nil, event.UserID), nil
		}
	}

	if fail := scanner.Err(); fail != nil {
		return nil, fail
	}

	return nil, errors.New("not found")
}

func (f *FileRepository) SaveMany(urls []*model.URL) error {
	file, err := os.OpenFile(f.fileStoragePath, os.O_RDWR|os.O_CREATE|os.O_APPEND, 0644)
	if err != nil {
		return err
	}

	defer file.Close()

	for _, u := range urls {
		event := newEvent(u.Original, u.Code)
		data, fail := json.Marshal(event)
		if fail != nil {
			return fail
		}

		if _, failed := file.Write(append(data, '\n')); failed != nil {
			return failed
		}

	}

	return nil
}

func (f *FileRepository) FindManyByUserID(userID string) ([]*model.URL, error) {
	file, err := os.Open(f.fileStoragePath)

	if err != nil {
		return nil, err
	}
	defer file.Close()

	scanner := bufio.NewScanner(file)

	var urls []*model.URL

	for scanner.Scan() {
		var event Event

		data := scanner.Bytes()
		err = json.Unmarshal(data, &event)
		if err != nil {
			continue
		}

		if event.UserID == userID {
			url := model.NewURL(event.ShortURL, event.OriginalURL, nil, userID)
			urls = append(urls, url)

			continue
		}

		if err = scanner.Err(); err != nil {
			return nil, err
		}

		return urls, nil
	}

	return urls, nil
}
