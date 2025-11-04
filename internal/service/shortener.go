package service

import (
	"crypto/sha256"
	"encoding/base64"
	"errors"
	"math/rand/v2"
	"net/url"
	"strconv"
	"strings"

	"github.com/yandex-practicum/shorten-url/internal/model"
	"github.com/yandex-practicum/shorten-url/internal/repository"
)

type ShortenerService struct {
	repo    repository.URLRepository
	BaseURL string
}

func NewShortenerService(repo repository.URLRepository, baseURL string) *ShortenerService {
	return &ShortenerService{repo: repo, BaseURL: baseURL}
}

func (s *ShortenerService) Shorten(original string) (*model.URL, error) {
	if ok, _ := s.isValidURL(original); !ok {
		return nil, errors.New("invalid URL")
	}

	code := s.HashURL(original)

	for {
		if u, err := s.repo.FindByCode(code); err == nil {
			if u.Original == original {
				return u, nil
			}

			original = original + strconv.Itoa(rand.Int())
			continue
		}

		u := model.New(code, original)

		if err := s.repo.Save(u); err != nil {
			return nil, err
		}

		return u, nil
	}
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

func (s *ShortenerService) isValidURL(original string) (bool, error) {
	u, err := url.ParseRequestURI(original)
	if err != nil {
		return false, err
	}

	if u.Scheme == " " || u.Host == "" {
		return false, nil
	}

	return true, nil
}
