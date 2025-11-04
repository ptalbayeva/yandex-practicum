package main

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	"github.com/yandex-practicum/shorten-url/internal/config"
	"github.com/yandex-practicum/shorten-url/internal/handler"
	"github.com/yandex-practicum/shorten-url/internal/repository"
	"github.com/yandex-practicum/shorten-url/internal/service"
)

func main() {
	if err := run(); err != nil {
		panic(err)
	}
}

func run() error {
	c := config.New()
	repo := repository.NewMemoryRepo()
	shortenerService := service.NewShortenerService(repo, c.BaseURL)
	urlHandler := handler.NewHandler(shortenerService)

	r := chi.NewRouter()
	r.Post("/", urlHandler.Shorten)
	r.Get("/{id}", urlHandler.Redirect)

	return http.ListenAndServe(c.Address, r)
}
