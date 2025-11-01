package handler

import (
	"io"
	"net/http"
	"net/http/httptest"
	"strings"
	"testing"

	"github.com/stretchr/testify/assert"
	"github.com/stretchr/testify/require"
	"github.com/yandex-practicum/shorten-url/internal/model"
	"github.com/yandex-practicum/shorten-url/internal/repository"
	"github.com/yandex-practicum/shorten-url/internal/service"
)

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
				response:    "http://localhost:8080/FgAJzmB",
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
			request := httptest.NewRequest(test.method, "http://localhost:8080/", strings.NewReader(test.url))
			request.Header.Set("Content-Type", test.contentType)
			w := httptest.NewRecorder()
			repo := repository.NewMemoryRepo()
			h := &Handler{
				shortener: service.NewShortenerService(repo),
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

			repo := repository.NewMemoryRepo()
			repo.Save(test.URL)
			require.NoError(t, repo.Save(test.URL))

			s := service.NewShortenerService(repo)

			mux := http.NewServeMux()
			mux.HandleFunc("/{id}", NewHandler(s).Redirect)

			request := httptest.NewRequest(http.MethodGet, "http://localhost:8080/"+test.code, nil)
			mux.ServeHTTP(w, request)

			res := w.Result()
			defer res.Body.Close()

			assert.Equal(t, test.want.code, res.StatusCode)

			location := res.Header.Get("Location")
			assert.Equal(t, test.want.location, location)
		})
	}
}
