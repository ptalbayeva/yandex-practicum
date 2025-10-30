package main

import (
	"net/http"

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
	repo := repository.NewMemoryRepo()
	shortenerService := service.NewShortenerService(repo)
	urlHandler := handler.NewHandler(shortenerService)

	mux := http.NewServeMux()
	mux.HandleFunc(`/`, urlHandler.Shorten)
	mux.HandleFunc(`/{id}`, urlHandler.Redirect)

	return http.ListenAndServe(`:8080`, mux)
}
