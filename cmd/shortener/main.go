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
	"time"

	_ "net/http/pprof"

	"github.com/go-chi/chi/v5"
	"github.com/go-chi/chi/v5/middleware"
	"github.com/golang-migrate/migrate/v4"
	"github.com/golang-migrate/migrate/v4/database/postgres"
	_ "github.com/golang-migrate/migrate/v4/source/file"
	_ "github.com/jackc/pgx/v5/stdlib"
	"github.com/yandex-practicum/shorten-url/internal/config"
	"github.com/yandex-practicum/shorten-url/internal/handler"
	g "github.com/yandex-practicum/shorten-url/internal/middleware"
	"github.com/yandex-practicum/shorten-url/internal/repository"
	"github.com/yandex-practicum/shorten-url/internal/service"
	"github.com/yandex-practicum/shorten-url/pkg/audit"
	"go.uber.org/zap"
)

var (
	BuildVersion string = "N/A"
	BuildDate    string = "N/A"
	BuildCommit  string = "N/A"
)

func main() {
	// ИНформация build-а
	log.Printf("Build version: %s\n", BuildVersion)
	log.Printf("Build date:    %s\n", BuildDate)
	log.Printf("Build commit:  %s\n", BuildCommit)
	log.Println()

	if err := run(); err != nil {
		g.Log.Error("Ошибка на сервере", zap.Error(err))
	}
}

func run() error {
	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	c := config.New()

	if err := g.Initialize(c.LogLevel); err != nil {
		return err
	}

	auditService, err := initAudit(c)

	if err != nil {
		return err
	}

	auditService.Start()

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
	urlHandler := handler.NewHandler(shortenerService, db, auditService)

	r := chi.NewRouter()
	r.Use(g.RequestLogger())
	r.Use(g.GzipMiddleware)
	r.Use(g.Auth([]byte(c.AuthKey)))

	r.Mount("/debug", middleware.Profiler())
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
		if c.EnableHttps {
			if err = service.EnsureCertificates(c.CertFile, c.KeyFile); err != nil {
				_ = fmt.Errorf("error while generating certificates %w", err)
			}

			log.Printf("Запуск HTTPS на %s", c.Address)
			err = server.ListenAndServeTLS(c.CertFile, c.KeyFile)
		} else {
			log.Printf("Запуск HTTP на %s", c.Address)
			err = server.ListenAndServe()
		}

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			_ = fmt.Errorf("error while starting server %w", err)
		}
	}()

	s := make(chan os.Signal, 1)
	signal.Notify(s, syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)

	<-s

	shutdownctx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	if fail := server.Shutdown(shutdownctx); fail != nil {
		_ = fmt.Errorf("server shutdown error: %w", fail)
	}

	return nil
}

func initAudit(config *config.Config) (audit.Publisher, error) {
	svc := audit.NewPublisherService(100)

	if config.AuditFile != "" {
		fileObserver, err := audit.NewFileObserver(config.AuditFile)
		if err != nil {
			return nil, err
		}
		svc.Subscribe(fileObserver)
	}

	if config.AuditURL != "" {
		httpObserver := audit.NewAPIObserver(config.AuditURL)
		svc.Subscribe(httpObserver)
	}

	return svc, nil
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
