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
				return u, repository.ErrConflict
			}

			original = original + strconv.Itoa(rand.Int())
			continue
		}

		u := model.NewURL(code, original, nil)

		if err := s.repo.Save(u); err != nil {
			if errors.Is(err, repository.ErrConflict) {
				return u, err
			}

			return nil, err
		}

		return u, nil
	}
}

func (s *ShortenerService) ShortenBatch(items []model.BatchURLRequest) ([]*model.BatchURLResponse, error) {
	responses := make([]*model.BatchURLResponse, 0, len(items))
	urls := make([]*model.URL, 0, len(items))

	for _, item := range items {
		if ok, _ := s.isValidURL(item.OriginalURL); !ok {
			return nil, errors.New("invalid URL")
		}

		code := s.HashURL(item.OriginalURL)

		for {
			if u, err := s.repo.FindByCode(code); err == nil {
				if u.Original == item.OriginalURL {
					urls = append(urls, u)
					responses = append(responses, &model.BatchURLResponse{
						CorrelationID: item.CorrelationID,
						ShortenURL:    s.BaseURL + "/" + u.Code,
					})

					break
				}

				item.OriginalURL = item.OriginalURL + strconv.Itoa(rand.Int())
				continue
			}

			u := model.NewURL(code, item.OriginalURL, item.CorrelationID)
			urls = append(urls, u)
			responses = append(responses, &model.BatchURLResponse{
				CorrelationID: item.CorrelationID,
				ShortenURL:    s.BaseURL + "/" + u.Code,
			})

			break
		}
	}

	if err := s.repo.SaveMany(urls); err != nil {
		return nil, err
	}

	return responses, nil
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
