package service

import (
	"context"
	"log"
	"time"

	"github.com/yandex-practicum/shorten-url/internal/model"
	"github.com/yandex-practicum/shorten-url/internal/repository"
)

type DeleteURLService struct {
	repository repository.URLRepository
	in         chan model.DeleteURLTask
}

func NewDeleteURLService(
	repository repository.URLRepository,
	bufferSize int,
) *DeleteURLService {
	return &DeleteURLService{repository: repository, in: make(chan model.DeleteURLTask, bufferSize)}
}

func (w *DeleteURLService) Enqueue(task model.DeleteURLTask) {
	w.in <- task
}

func (w *DeleteURLService) Run(ctx context.Context) {
	const (
		maxBatchSize = 100
		flushTimeout = 500 * time.Millisecond
	)

	ticker := time.NewTicker(flushTimeout)
	defer ticker.Stop()

	buffer := make([]model.DeleteURLTask, 0, maxBatchSize)

	flush := func() {
		if len(buffer) == 0 {
			return
		}

		grouped := make(map[string][]string)
		for _, task := range buffer {
			grouped[task.UserID] = append(grouped[task.UserID], task.Short)
		}

		for userID, codes := range grouped {
			if err := w.repository.DeleteManyByCodes(userID, codes); err != nil {
				log.Println("could not delete user URLs:" + err.Error())
			}
		}

		buffer = buffer[:0]
	}

	for {
		select {
		case <-ctx.Done():
			flush()
			return

		case task := <-w.in:
			buffer = append(buffer, task)

			if len(buffer) >= maxBatchSize {
				flush()
			}

		case <-ticker.C:
			flush()
		}
	}
}
