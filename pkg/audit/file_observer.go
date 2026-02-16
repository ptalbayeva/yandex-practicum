package audit

import (
	"encoding/json"
	"os"
)

type FileObserver struct {
	file *os.File
}

func NewFileObserver(filepath string) (error, *FileObserver) {
	file, err := os.OpenFile(filepath, os.O_APPEND|os.O_CREATE|os.O_WRONLY, 0644)
	if err != nil {
		return err, nil
	}

	defer file.Close()

	return nil, &FileObserver{file: file}
}

func (f FileObserver) Publish(event Event) error {
	data, err := json.Marshal(event)
	if err != nil {
		return err
	}

	_, err = f.file.Write(append(data, '\n'))

	return err
}
