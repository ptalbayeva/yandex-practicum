package repository

import (
	"errors"

	"github.com/yandex-practicum/shorten-url/internal/model"
)

var ErrConflict = errors.New("conflict")

type URLRepository interface {
	Save(u *model.URL) error
	FindByCode(code string) (*model.URL, error)
	SaveMany(u []*model.URL) error
	FindManyByUserID(userID string) ([]*model.URL, error)
	DeleteManyByCodes(userID string, codes []string) error
}
