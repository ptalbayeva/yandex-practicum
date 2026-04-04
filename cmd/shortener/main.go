package main

import (
	"context"
	"database/sql"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os/signal"
	"sync"
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
	"github.com/yandex-practicum/shorten-url/pkg/proto"
	"go.uber.org/zap"
	"google.golang.org/grpc"
	"google.golang.org/grpc/reflection"
)

var (
	BuildVersion string = "N/A"
	BuildDate    string = "N/A"
	BuildCommit  string = "N/A"
)

func main() {
	// ИНформация build-а
	g.Log.Info("Build info: ",
		zap.String("Build version", BuildVersion),
		zap.String("Build date", BuildDate),
		zap.String("Commit SHA", BuildCommit),
	)

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

	if c.TrustedSubnet != "" {
		r.Use(g.TrustedSubnetMiddleware(c))
	}

	r.Mount("/debug", middleware.Profiler())
	r.Post("/", urlHandler.Shorten)
	r.Get("/{id}", urlHandler.ExpandURL)
	r.Get("/api/user/urls", urlHandler.ListUserURLs)
	r.Post("/api/shorten", urlHandler.ShortenURL)
	r.Post("/api/shorten/batch", urlHandler.BatchShorten)
	r.Delete("/api/user/urls", urlHandler.DeleteUserURLs)
	r.Get("/api/internal/stats", urlHandler.GetInternalStats)
	r.Get("/ping", urlHandler.Ping)

	grpcListen, err := net.Listen("tcp", c.GRPCAddr)
	if err != nil {
		g.Log.Fatal("failed to listen gRPC", zap.Error(err))
	}

	authInterceptor := g.AuthGRPCInterceptor([]byte(c.AuthKey))
	grpcServer := grpc.NewServer(grpc.UnaryInterceptor(authInterceptor))
	shortenerServer := service.NewShortenerServer(shortenerService)
	proto.RegisterShortenerServiceServer(grpcServer, shortenerServer)
	reflection.Register(grpcServer)

	ctx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM, syscall.SIGQUIT)
	defer stop()

	var wg sync.WaitGroup

	wg.Add(1)
	go func() {
		defer wg.Done()
		g.Log.Info("Запуск gRPC на %s", zap.String("address", c.GRPCAddr))

		if err = grpcServer.Serve(grpcListen); err != nil && !errors.Is(err, grpc.ErrServerStopped) {
			g.Log.Error("gRPC server failed", zap.Error(err))
		}
	}()

	server := &http.Server{
		Addr:    c.Address,
		Handler: r,
	}

	wg.Add(1)
	go func() {
		defer wg.Done()
		var err error
		if c.EnableHttps {
			g.Log.Info("Запуск HTTPS на %s", zap.String("address", c.Address))
			err = server.ListenAndServeTLS(c.CertFile, c.KeyFile)
		} else {
			g.Log.Info("Запуск HTTP на %s", zap.String("address", c.Address))
			err = server.ListenAndServe()
		}

		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			g.Log.Error("HTTP server error", zap.Error(err))
		}
	}()

	<-ctx.Done()
	g.Log.Info("Получен сигнал завершения, начинаем shutdown...")

	shutdownCtx, cancel := context.WithTimeout(context.Background(), 30*time.Second)
	defer cancel()

	go func() {
		if err = server.Shutdown(shutdownCtx); err != nil {
			g.Log.Error("HTTP shutdown error", zap.Error(err))
		}
		grpcServer.GracefulStop()
	}()

	wg.Wait()
	g.Log.Info("Все сервисы успешно остановлены")

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
