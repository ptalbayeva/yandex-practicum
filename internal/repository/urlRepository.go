package repository

import "github.com/yandex-practicum/shorten-url/internal/model"

type URLRepository interface {
	Save(u *model.URL) error
	FindByCode(code string) (*model.URL, error)
	SaveMany(u []*model.URL) error
}
