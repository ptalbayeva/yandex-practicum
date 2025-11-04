package repository

import (
	"errors"
	"sync"

	"github.com/yandex-practicum/shorten-url/internal/model"
)

type MemoryRepo struct {
	mu   sync.RWMutex
	data map[string]*model.URL
}

func NewMemoryRepo() *MemoryRepo {
	return &MemoryRepo{
		data: make(map[string]*model.URL),
	}
}

func (r *MemoryRepo) Save(u *model.URL) error {
	r.mu.RLock()
	defer r.mu.RUnlock()

	r.data[u.Code] = u

	return nil
}

func (r *MemoryRepo) FindByCode(code string) (*model.URL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()

	u, ok := r.data[code]

	if !ok {
		return nil, errors.New("not found")
	}

	return u, nil
}
