package main

import (
	"context"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/yandex-practicum/shorten-url/internal/config"
	"github.com/yandex-practicum/shorten-url/internal/handler"
	"github.com/yandex-practicum/shorten-url/internal/middleware"
	"github.com/yandex-practicum/shorten-url/internal/repository"
	"github.com/yandex-practicum/shorten-url/internal/service"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		middleware.Log.Fatal("Ошибка на сервере", zap.Error(err))
	}
}

func run() error {
	c := config.New()

	if err := middleware.Initialize(c.LogLevel); err != nil {
		return err
	}

	repo := repository.NewMemoryRepo()
	shortenerService := service.NewShortenerService(repo, c.BaseURL)
	urlHandler := handler.NewHandler(shortenerService)

	r := chi.NewRouter()
	r.Post("/", urlHandler.Shorten)
	r.Get("/{id}", urlHandler.Redirect)
	r.Post("/api/shorten", urlHandler.ShortenJSON)

	server := &http.Server{
		Addr:    c.Address,
		Handler: middleware.RequestLogger(middleware.GzipHandler(r)),
	}

	go func() {
		if err := server.ListenAndServe(); err != nil {
			log.Fatal(err)
		}
	}()

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)

	<-s
	if err := server.Shutdown(context.Background()); err != nil {
		log.Fatal(err)
	}

	return nil
}
