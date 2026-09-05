package main

import (
	"cbr-worker/internal"
	"cbr-worker/internal/cbr"
	"context"
	"net/http"

	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

const (
	dbConnTimeout  = 10 * time.Second
	collectTimeout = 30 * time.Second
)

func run() int {
	_ = godotenv.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	cfg, err := pgxpool.ParseConfig(os.Getenv("DATABASE_URL"))
	if err != nil {
		logger.Error("Failed to parse database config. Exiting...", slog.Any("error", err))
		return 1
	}

	defCtx := context.Background()
	ctx, cancel := context.WithTimeout(defCtx, dbConnTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		logger.Error("Failed to create database connection pool. Exiting...", slog.Any("error", err))
		return 2
	}
	defer pool.Close()

	httpClient := &http.Client{
		Timeout: time.Second * 30,
	}
	cbrClient := cbr.NewClient(httpClient,
		logger.With(slog.String("component", "client")),
	)
	collector := internal.NewCollector(cbrClient, pool,
		logger.With(slog.String("component", "collector")),
	)

	logger.Info("Starting collector...")
	ctx, cancel = context.WithTimeout(defCtx, collectTimeout)
	defer cancel()

	err = collector.Collect(ctx)
	if err != nil {
		logger.Error("Failed to collect exchange rates", slog.Any("error", err))
		return 4
	}

	logger.Info("Collector succeeded")
	return 0
}

func main() {
	os.Exit(run())
}
