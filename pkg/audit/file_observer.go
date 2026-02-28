package audit

import (
	"encoding/json"
	"os"
)

// FileObserver файловый наблюдатель
type FileObserver struct {
	file *os.File
}

// NewFileObserver создание файлового наблюдателя
func NewFileObserver(filepath string) (*FileObserver, error) {
	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return nil, err
	}

	defer file.Close()

	return &FileObserver{file: file}, nil
}

// Publish публикует события
func (f FileObserver) Publish(event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, err = f.file.Write(append(data, '\n'))

	return err
}
