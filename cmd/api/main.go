package main

import (
	"cbr-worker/internal/cbr"
	"cbr-worker/internal/http"
	"context"
	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

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

	ctx, cancel := context.WithTimeout(context.Background(), dbConnTimeout)
	defer cancel()

	pool, err := pgxpool.NewWithConfig(ctx, cfg)
	if err != nil {
		logger.Error("Failed to create database connection pool. Exiting...", slog.Any("error", err))
		return 2
	}
	defer pool.Close()

	repo := cbr.NewRepository(pool,
		logger.With(slog.String("component", "repository")),
	)

	srv := http.NewServer(httpServerAddr, repo,
		logger.With(slog.String("component", "http-server")),
	)

	if err := srv.Start(context.TODO()); err != nil {
		logger.Error("Failed to start server. Exiting...", slog.Any("error", err))

		return 3
	}

	return 0
}

func main() {
	os.Exit(run())
}

// TODO: rename endpoint and add query params
