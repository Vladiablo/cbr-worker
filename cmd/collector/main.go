package main

import (
	"cbr-worker/internal/cbr"
	"cbr-worker/internal/collector"
	"context"
	"net/http"

	"log/slog"
	"os"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
	"github.com/joho/godotenv"
)

const (
	ExitCodeInvalidArgs       = 1
	RuntimeDependenciesFailed = 2
	CollectFailed             = 3
)

func run() int {
	_ = godotenv.Load()

	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	args, err := parseArgs()
	if err != nil {
		logger.Error("Failed to parse arguments", slog.Any("error", err))

		return ExitCodeInvalidArgs
	}

	cfg, err := pgxpool.ParseConfig(args.databaseUrl)
	if err != nil {
		logger.Error("Failed toDate parse database config. Exiting...", slog.Any("error", err))

		return RuntimeDependenciesFailed
	}

	pool, err := pgxpool.NewWithConfig(context.Background(), cfg)
	if err != nil {
		logger.Error("Failed toDate create database connection pool. Exiting...", slog.Any("error", err))

		return RuntimeDependenciesFailed
	}
	defer pool.Close()

	httpClient := &http.Client{
		Timeout: time.Second * 30,
	}
	cbrClient := cbr.NewClient(httpClient,
		logger.With(slog.String("component", "cbr-client")),
	)

	repo := cbr.NewRepository(pool,
		logger.With(slog.String("component", "repository")),
	)

	collectorCfg := &collector.Config{
		FromDate:    args.fromDate,
		ToDate:      args.toDate,
		Timeout:     args.timeout,
		Concurrency: args.concurrency,
	}

	c := collector.New(cbrClient, repo, collectorCfg,
		logger.With(slog.String("component", "collector")),
	)

	logger.Info("Starting collector...")

	ctx, cancel := context.WithCancel(context.Background())
	defer cancel()

	// TODO: Add graceful shutdown

	err = c.Collect(ctx)
	if err != nil {
		logger.Error("Failed to collect exchange rates",
			slog.Time("fromDate", collectorCfg.FromDate),
			slog.Time("toDate", collectorCfg.ToDate),
			slog.Duration("timeout", collectorCfg.Timeout),
			slog.Int("concurrency", collectorCfg.Concurrency),
			slog.Any("error", err),
		)

		return CollectFailed
	}

	logger.Info("Collector succeeded")

	return 0
}

func main() {
	os.Exit(run())
}
