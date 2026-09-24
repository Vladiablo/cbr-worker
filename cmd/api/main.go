package main

import (
	"cbr-worker/internal/cbr/repository"
	"cbr-worker/internal/cbr/service"
	"cbr-worker/internal/http"
	"context"
	"log/slog"
	"os"
	"os/signal"
	"sync"
	"sync/atomic"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

//	@title			CBR Worker API
//	@version		1.0
//	@description	API to get CBR exchange rates

//	@host		localhost:8080
//	@BasePath	/v1

//	@externalDocs.description	OpenAPI
//	@externalDocs.url			https://swagger.io/resources/open-api/

const (
	httpServerAddr = ":8080"
	dbConnTimeout  = 10 * time.Second
)

func run() int {
	_ = godotenv.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	cfg, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		logger.Error("Failed to parse database config. Exiting...", slog.Any("error", err))
		return 1
	}

	var pool *pgxpool.Pool
	{
		ctx, cancel := context.WithTimeout(context.Background(), dbConnTimeout)
		defer cancel()

		pool, err = pgxpool.NewWithConfig(ctx, cfg)
		if err != nil {
			logger.Error("Failed to create database connection pool. Exiting...", slog.Any("error", err))
			return 2
		}
		defer pool.Close()
	}

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	repo := repository.NewPgRepository(pool,
		logger.With(slog.String("component", "pg-repo")),
	)

	cachedRepoCfg := &repository.CachedRepositoryConfig{
		CacheUpdateInterval: 5 * time.Minute,
	}
	cachedRepo := repository.NewCachedRepository(repo, cachedRepoCfg,
		logger.With(slog.String("component", "cached-repo")),
	)

	if err := cachedRepo.Init(ctx); err != nil {
		logger.Error("Failed to init cached repository. Exiting...", slog.Any("error", err))

		return 1
	}

	svc := service.New(cachedRepo)

	srv := http.NewServer(httpServerAddr, svc,
		logger.With(slog.String("component", "http-server")),
	)

	var wg sync.WaitGroup
	wg.Add(1)

	var firstErr atomic.Pointer[error]

	go func() {
		defer wg.Done()
		defer cancel()

		if err := srv.Start(ctx); err != nil {
			firstErr.CompareAndSwap(nil, &err)

			logger.Error("Failed to serve HTTP", slog.Any("error", err))
		}
	}()

	doneCh := make(chan struct{})
	go func() {
		defer close(doneCh)

		wg.Wait()
	}()

	// Set up channel on which to send signal notifications.
	// We must use a buffered channel or risk missing the signal
	// if we're not ready to receive when the signal is sent.
	sigCh := make(chan os.Signal, 1)
	signal.Notify(sigCh, os.Interrupt)

	var timeoutCh <-chan time.Time
	const shutdownTimeout = 10 * time.Second

	ctxDoneCh := ctx.Done()

	for {
		select {
		case sig := <-sigCh:
			logger.Info("Received signal. Shutting down...", slog.String("signal", sig.String()))

			cancel()

		case <-ctxDoneCh:
			ctxDoneCh = nil
			timeoutCh = time.After(shutdownTimeout)

		case <-timeoutCh:
			logger.Error("Failed to graceful shutdown", slog.String("duration", shutdownTimeout.String()))

			return 1

		case <-doneCh:
			if firstErr.Load() != nil {
				return 3
			}

			logger.Info("Graceful shutdown complete")
			return 0
		}
	}
}

func main() {
	os.Exit(run())
}
