package service

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"strings"

	"github.com/yandex-practicum/shorten-url/internal/model"
	"github.com/yandex-practicum/shorten-url/internal/repository"
)

type ShortenerService struct {
	repo    repository.URLRepository
	baseURL string
}

func NewShortenerService(repo repository.URLRepository, baseURL string) *ShortenerService {
	return &ShortenerService{repo: repo, baseURL: baseURL}
}

func (s *ShortenerService) Shorten(original string) (*model.URL, error) {
	code := s.HashURL(original)
	u := model.New(code, original)

	if err := s.repo.Save(u); err != nil {
		return nil, err
	}

	return u, nil
}

func (s *ShortenerService) Resolve(code string) (*model.URL, error) {
	u, err := s.repo.FindByCode(code)

	if err != nil {
		return nil, errors.New("not found")
	}

	return u, nil
}

func (s *ShortenerService) HashURL(original string) string {
	hash := sha256.Sum256([]byte(original))
	encoded := base64.URLEncoding.EncodeToString(hash[:])

	return strings.TrimRight(encoded, "=")[:7]
}
