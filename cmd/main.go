package main

import (
	"cbr-worker/internal/cbr"
	"cbr-worker/internal/http/handlers"
	"log/slog"
	"net/http"
	"os"
	"time"
)

const httpServerAddr = ":8080"

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))

	httpClient := &http.Client{
		Timeout: time.Second * 20,
	}
	cbrClient := cbr.NewClient(httpClient,
		logger.With(slog.String("component", "client")),
	)

	currencyHandler := handlers.NewCurrencyHandler(cbrClient,
		logger.With(slog.String("component", "http")),
	)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/getRates", currencyHandler.GetRates)

	srv := &http.Server{
		Addr:         httpServerAddr,
		Handler:      mux,
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	logger.Info("Starting HTTP server...", slog.String("addr", httpServerAddr))
	err := srv.ListenAndServe()
	if err != nil {
		logger.Error("Failed to start HTTP server. Exiting...", slog.Any("error", err))
		os.Exit(1)
	}
}
