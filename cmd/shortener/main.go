package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"log"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/go-chi/chi/v5"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
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

	db, err := sql.Open("pgx", c.DatabaseDSN)
	if err != nil {
		return err
	}

	defer db.Close()

	repo, err := initRepository(*c)
	if err != nil {
		return err
	}

	fail := applyMigrations(db, "./migrations")
	if fail != nil {
		return err
	}

	shortenerService := service.NewShortenerService(repo, c.BaseURL)
	urlHandler := handler.NewHandler(shortenerService, db)

	r := chi.NewRouter()
	r.Use(middleware.RequestLogger())
	r.Use(middleware.GzipHandler())

	r.Post("/", urlHandler.Shorten)
	r.Get("/{id}", urlHandler.Redirect)
	r.Post("/api/shorten", urlHandler.ShortenJSON)
	r.Get("/ping", urlHandler.Ping)

	server := &http.Server{
		Addr:    c.Address,
		Handler: r,
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

func initRepository(cfg config.Config) (repository.URLRepository, error) {
	if cfg.DatabaseDSN != "" {
		db, err := sql.Open("pgx", cfg.DatabaseDSN)
		if err != nil {
			return nil, err
		}

		defer db.Close()

		return repository.NewDBRepository(db), nil
	}

	if cfg.FileStoragePath != "" {
		return repository.NewFileRepository(cfg.FileStoragePath), nil
	}

	return repository.NewMemoryRepo(), nil
}

func applyMigrations(db *sql.DB, migrationsPath string) error {
	driver, err := postgres.WithInstance(db, &postgres.Config{})
	if err != nil {
		return fmt.Errorf("failed to create migration driver: %w", err)
	}

	m, err := migrate.NewWithDatabaseInstance(
		fmt.Sprintf("file://%s", migrationsPath),
		"postgres",
		driver,
	)
	if err != nil {
		return fmt.Errorf("failed to create migrate instance: %w", err)
	}

	err = m.Up()
	if err != nil && !errors.Is(err, migrate.ErrNoChange) {
		return fmt.Errorf("failed to apply migrations: %w", err)
	}

	return nil
}
