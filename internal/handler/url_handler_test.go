package handler

import (
	"database/sql"
	"io"
	"net/http"
	"net/http/httptest"
	"os"
	"strings"
	"testing"

	"github.com/go-chi/chi/v5"
	"github.com/go-resty/resty/v2"
	"github.com/google/uuid"
	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yandex-practicum/shorten-url/internal/config"
	"github.com/yandex-practicum/shorten-url/internal/model"
	"github.com/yandex-practicum/shorten-url/internal/repository"
	"github.com/yandex-practicum/shorten-url/internal/service"
	"github.com/yandex-practicum/shorten-url/pkg/audit"
)

var testC *config.Config

func TestMain(m *testing.M) {
	testC = &config.Config{
		Address:     "localhost:8081",
		BaseURL:     "http://localhost:8081",
		DatabaseDSN: "test.db",
	}

	code := m.Run()
	os.Exit(code)
}

func Test_Shorten(t *testing.T) {
	type want struct {
		code        int
		response    string
		contentType string
	}

	tests := []struct {
		name        string
		url         string
		method      string
		contentType string
		want        want
	}{
		{
			"Позитивный кейс",
			"https://yandex.ru",
			http.MethodPost,
			"text/plain",
			want{
				code:        http.StatusCreated,
				response:    "http://localhost:8081/FgAJzmB",
				contentType: "text/plain",
			},
		},
		{
			"Передача json",
			"{https://yandex.ru}",
			http.MethodPost,
			"application/json",
			want{
				code:        http.StatusBadRequest,
				response:    "content type must be text/plain\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			"Невалидный метод",
			"",
			http.MethodGet,
			"text/plain",
			want{
				code:        http.StatusMethodNotAllowed,
				response:    "method not allowed\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
		{
			"Пустой url был передан",
			"",
			http.MethodPost,
			"text/plain",
			want{
				code:        http.StatusBadRequest,
				response:    "missing url\n",
				contentType: "text/plain; charset=utf-8",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			request := httptest.NewRequest(test.method, testC.BaseURL, strings.NewReader(test.url))
			request.Header.Set("Content-Type", test.contentType)
			w := httptest.NewRecorder()
			repo := repository.NewMemoryRepo()
			deleteURL := service.NewDeleteURLService(repo, 100)
			h := &Handler{
				shortener:    service.NewShortenerService(repo, testC.BaseURL, *deleteURL),
				auditService: audit.NewNoopPublisher(),
			}
			h.Shorten(w, request)

			res := w.Result()
			assert.Equal(t, test.want.code, res.StatusCode)

			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)

			require.NoError(t, err)
			assert.Equal(t, test.want.response, string(resBody))
			assert.Equal(t, test.want.contentType, res.Header.Get("Content-Type"))
		})
	}
}
func TestHandler_ShortenURL(t *testing.T) {
	repo := repository.NewMemoryRepo()
	deleteURL := service.NewDeleteURLService(repo, 100)
	handler := &Handler{
		shortener: service.NewShortenerService(
			repo,
			testC.BaseURL,
			*deleteURL,
		),
	}
	h := http.HandlerFunc(handler.ShortenURL)
	srv := httptest.NewServer(h)

	type want struct {
		code     int
		response *model.Response
	}

	tests := []struct {
		name        string
		method      string
		contentType string
		request     *model.Request
		want        want
	}{
		{
			name:        "Позитивный кейс",
			method:      http.MethodPost,
			contentType: "application/json",
			request:     &model.Request{URL: "{\n \"url\":\"https://practicum.yandex.ru\"\n}"},
			want: want{
				code:     http.StatusCreated,
				response: &model.Response{Result: "{\n \"result\": \"http://localhost:8081/ipkjUVt\"\n}"},
			},
		},
		{
			name:        "Некорректный метод",
			method:      http.MethodGet,
			contentType: "",
			request:     &model.Request{URL: "{\n \"url\":\"https://practicum.yandex.ru\"\n}"},
			want: want{
				code:     http.StatusMethodNotAllowed,
				response: nil,
			},
		},
		{
			name:        "Некорректный content-type",
			method:      http.MethodPost,
			contentType: "text/plain",
			request:     &model.Request{URL: "https://practicum.yandex.ru"},
			want: want{
				code:     http.StatusBadRequest,
				response: nil,
			},
		},
		{
			name:        "Пустое тело",
			method:      http.MethodPost,
			contentType: "application/json",
			request:     &model.Request{URL: "{}"},
			want: want{
				code:     http.StatusUnprocessableEntity,
				response: nil,
			},
		},
	}
	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			var r io.Reader
			if tt.request != nil {
				r = strings.NewReader(tt.request.URL)
			}

			req := httptest.NewRequest(tt.method, srv.URL, r)
			req.Header.Add("Content-Type", tt.contentType)

			w := httptest.NewRecorder()
			h.ServeHTTP(w, req)

			res := w.Result()
			defer res.Body.Close()
			resBody, err := io.ReadAll(res.Body)
			assert.NoError(t, err)

			assert.Equal(t, tt.want.code, res.StatusCode)
			if tt.want.response != nil {
				assert.JSONEq(t, tt.want.response.Result, string(resBody))
			}
		})
	}
}

func Test_Redirect(t *testing.T) {
	type want struct {
		code     int
		location string
	}

	tests := []struct {
		name string
		code string
		URL  *model.URL
		want want
	}{
		{
			name: "Позитивный кейс",
			code: "FgAJzmB",
			URL: &model.URL{
				Code:     "FgAJzmB",
				Original: "https://yandex.ru",
			},
			want: want{
				http.StatusTemporaryRedirect,
				"https://yandex.ru",
			},
		},
		{
			name: "Не нашли по коду url",
			code: "1234rt",
			URL:  &model.URL{},
			want: want{
				http.StatusNotFound,
				"",
			},
		},
	}

	for _, test := range tests {
		t.Run(test.name, func(t *testing.T) {
			w := httptest.NewRecorder()
			r := getTestRouter(t, test.URL)

			request, err := http.NewRequest(http.MethodGet, "/"+test.code, nil)
			require.NoError(t, err)

			r.ServeHTTP(w, request)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.code, res.StatusCode)

			location := res.Header.Get("Location")
			assert.Equal(t, test.want.location, location)
		})
	}
}

func getTestRouter(t *testing.T, url *model.URL) chi.Router {
	r := chi.NewRouter()

	repo := repository.NewMemoryRepo()
	repo.Save(url)
	require.NoError(t, repo.Save(url))

	deleteURL := service.NewDeleteURLService(repo, 100)
	s := service.NewShortenerService(repo, testC.BaseURL, *deleteURL)
	noopAudit := audit.NewNoopPublisher()
	handler := http.HandlerFunc(NewHandler(s, &sql.DB{}, noopAudit).ExpandURL)

	r.Get("/{id}", handler)

	return r
}

func TestHandler_BatchShorten(t *testing.T) {
	type shortenBatchRequest struct {
		CorrelationID string `json:"correlation_id"`
		OriginalURL   string `json:"original_url"`
	}

	type shortenBatchResponse struct {
		CorrelationID string `json:"correlation_id"`
		ShortURL      string `json:"short_url"`
	}

	var responseData []shortenBatchResponse

	requestData := []shortenBatchRequest{
		{
			CorrelationID: uuid.NewString(),
			OriginalURL:   "https://practicum.yandex.ru",
		},
		{
			CorrelationID: uuid.NewString(),
			OriginalURL:   "https://yandex.ru",
		},
	}

	repo := repository.NewMemoryRepo()
	deleteURL := service.NewDeleteURLService(repo, 100)
	s := service.NewShortenerService(repo, testC.BaseURL, *deleteURL)
	handler := &Handler{shortener: s}
	h := http.HandlerFunc(handler.BatchShorten)
	srv := httptest.NewServer(h)

	rest := resty.New().SetBaseURL(srv.URL)
	req := rest.R().
		SetHeader("Content-Type", "application/json").
		SetBody(requestData).
		SetResult(&responseData)

	resp, err := req.Post("/api/shorten/batch")

	assert.NoError(t, err)
	assert.Equal(t, http.StatusCreated, resp.StatusCode())
	assert.Equal(t, len(requestData), len(responseData))

	var correlations int
	for _, r := range responseData {
		for _, request := range requestData {
			if r.CorrelationID == request.CorrelationID {
				correlations++
			}
		}
	}

	assert.Equal(t, correlations, len(requestData))
}
