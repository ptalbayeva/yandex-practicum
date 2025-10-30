package repository

import (
	"errors"

	"github.com/yandex-practicum/shorten-url/internal/model"
)

type MemoryRepo struct {
	data map[string]*model.URL
}

func NewMemoryRepo() *MemoryRepo {
	return &MemoryRepo{
		data: make(map[string]*model.URL),
	}
}

func (r *MemoryRepo) Save(u *model.URL) error {
	r.data[u.Code] = u

	return nil
}

func (r *MemoryRepo) FindByCode(code string) (*model.URL, error) {
	u, ok := r.data[code]

	if !ok {
		return nil, errors.New("not found")
	}

	return u, nil
}
