package service

import (
	"context"
	"time"

	g "github.com/yandex-practicum/shorten-url/internal/middleware"
	"github.com/yandex-practicum/shorten-url/internal/model"
	"github.com/yandex-practicum/shorten-url/internal/repository"
	"go.uber.org/zap"
)

// DeleteURLService сервис для удаления url
type DeleteURLService struct {
	repository repository.URLRepository
	in         chan model.DeleteURLTask
}

// NewDeleteURLService создание сервиса
func NewDeleteURLService(
	repository repository.URLRepository,
	bufferSize int,
) *DeleteURLService {
	return &DeleteURLService{repository: repository, in: make(chan model.DeleteURLTask, bufferSize)}
}

// Enqueue добавление таски в очередь
func (w *DeleteURLService) Enqueue(task model.DeleteURLTask) {
	w.in <- task
}

// Run удаляет url
func (w *DeleteURLService) Run(ctx context.Context) {
	const (
		maxBatchSize = 100
		flushTimeout = 500 * time.Millisecond
	)

	ticker := time.NewTicker(flushTimeout)
	defer ticker.Stop()

	buffer := make([]model.DeleteURLTask, 0, maxBatchSize)

	flush := func() {
		defer func() {
			if r := recover(); r != nil {
				g.Log.Error("panic in delete:", zap.Any("recover return", r))
			}
		}()

		if len(buffer) == 0 {
			return
		}

		grouped := make(map[string][]string)
		for _, task := range buffer {
			grouped[task.UserID] = append(grouped[task.UserID], task.Short)
		}

		for userID, codes := range grouped {
			if err := w.repository.DeleteManyByCodes(userID, codes); err != nil {
				g.Log.Error("could not delete user URLs:", zap.Error(err))
			}
		}

		buffer = buffer[:0]
	}

	for {
		select {
		case <-ctx.Done():
			for {
				select {
				case task, ok := <-w.in:
					if !ok {
						flush()
						return
					}
					buffer = append(buffer, task)
					if len(buffer) >= maxBatchSize {
						flush()
					}
				default:
					flush()
					return
				}
			}

		case task, ok := <-w.in:
			if !ok {
				flush()
				return
			}

			buffer = append(buffer, task)
			if len(buffer) >= maxBatchSize {
				flush()
			}

		case <-ticker.C:
			flush()
		}
	}
}
