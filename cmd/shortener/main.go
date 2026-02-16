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
	g "github.com/yandex-practicum/shorten-url/internal/middleware"
	"github.com/yandex-practicum/shorten-url/internal/repository"
	"github.com/yandex-practicum/shorten-url/internal/service"
	"go.uber.org/zap"
)

func main() {
	if err := run(); err != nil {
		g.Log.Fatal("Ошибка на сервере", zap.Error(err))
	}
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := config.New()

	if err := g.Initialize(c.LogLevel); err != nil {
		return err
	}

	db, err := sql.Open("pgx", c.DatabaseDSN)
	if err != nil {
		return err
	}

	defer db.Close()

	repo, callback, err := initRepository(*c)
	if err != nil {
		return err
	}

	defer callback()

	deleteURLService := service.NewDeleteURLService(repo, 100)
	go deleteURLService.Run(ctx)

	shortenerService := service.NewShortenerService(repo, c.BaseURL, *deleteURLService)
	urlHandler := handler.NewHandler(shortenerService, db)

	r := chi.NewRouter()
	r.Use(g.RequestLogger())
	r.Use(g.GzipMiddleware)
	r.Use(g.Auth([]byte(c.AuthKey)))

	r.Post("/", urlHandler.Shorten)
	r.Get("/{id}", urlHandler.Redirect)
	r.Get("/api/user/urls", urlHandler.GetURLS)
	r.Post("/api/shorten", urlHandler.ShortenJSON)
	r.Post("/api/shorten/batch", urlHandler.BatchShorten)
	r.Delete("/api/user/urls", urlHandler.DeleteUserURLs)
	r.Get("/ping", urlHandler.Ping)

	server := &http.Server{
		Addr:    c.Address,
		Handler: r,
	}

	go func() {
		if fail := server.ListenAndServe(); fail != nil && !errors.Is(fail, http.ErrServerClosed) {
			log.Fatalf("server listen error: %v", fail)
		}
	}()

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM)

	<-s

	if fail := server.Shutdown(context.Background()); fail != nil {
		log.Fatalf("server shutdown error: %v", fail)
	}

	return nil
}

func initRepository(cfg config.Config) (repository.URLRepository, func(), error) {
	if cfg.DatabaseDSN != "" {
		db, err := sql.Open("pgx", cfg.DatabaseDSN)
		if err != nil {
			return nil, func() {}, err
		}

		err = applyMigrations(db, "./migrations")
		if err != nil {
			return nil, func() {}, err
		}

		repo := repository.NewDBRepository(db)

		return repo, func() { repo.Close() }, nil
	}

	if cfg.FileStoragePath != "" {
		return repository.NewFileRepository(cfg.FileStoragePath), func() {}, nil
	}

	return repository.NewMemoryRepo(), func() {}, nil
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
