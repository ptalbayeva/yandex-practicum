package repository

import (
	"context"
	"errors"

	"github.com/yandex-practicum/shorten-url/internal/model"
)

// ErrConflict ошибка при дубликате записи
var ErrConflict = errors.New("conflict")

// URLRepository интерфейс репозитория для работы с url
type URLRepository interface {
	Save(u *model.URL) error
	FindByCode(code string) (*model.URL, error)
	SaveMany(u []*model.URL) error
	FindManyByUserID(userID string) ([]*model.URL, error)
	DeleteManyByCodes(userID string, codes []string) error
	FindTotalURLs(ctx context.Context) (int, error)
	FindTotalUserIDs(ctx context.Context) (int, error)
}
