package repository

import (
	"errors"
	"slices"
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
	r.mu.Lock()
	defer r.mu.Unlock()

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

func (r *MemoryRepo) SaveMany(urls []*model.URL) error {
	r.mu.Lock()
	for _, u := range urls {
		r.data[u.Code] = u
	}

	defer r.mu.Unlock()

	return nil
}

func (r *MemoryRepo) FindManyByUserID(userID string) ([]*model.URL, error) {
	r.mu.RLock()
	defer r.mu.RUnlock()
	var urls []*model.URL

	return urls, nil
}

func (r *MemoryRepo) DeleteManyByCodes(userID string, codes []string) error {
	r.mu.Lock()
	defer r.mu.Unlock()

	for _, url := range r.data {
		if !url.IsDeleted && url.UserID == userID && slices.Contains(codes, url.Code) {
			url.IsDeleted = true
		}
	}

	return nil
}
