package repository

import "github.com/yandex-practicum/shorten-url/internal/model"

type UrlRepository interface {
	Save(u *model.URL) error
	FindByCode(code string) (*model.URL, error)
}
