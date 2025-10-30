package service

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"fmt"
	"strings"

	"github.com/yandex-practicum/shorten-url/internal/model"
	"github.com/yandex-practicum/shorten-url/internal/repository"
)

type ShortenerService struct {
	repo repository.UrlRepository
}

func NewShortenerService(repo repository.UrlRepository) *ShortenerService {
	return &ShortenerService{repo: repo}
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
		return nil, errors.New(fmt.Sprintf("shorten code %s not found", code))
	}

	return u, nil
}

func (s *ShortenerService) HashURL(original string) string {
	hash := sha256.Sum256([]byte(original))
	encoded := base64.URLEncoding.EncodeToString(hash[:])

	return strings.TrimRight(encoded, "=")[:7]
}
