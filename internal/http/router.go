package http

import (
	"cbr-worker/internal/cbr"
	"cbr-worker/internal/http/handlers"
	"log/slog"
	"net/http"
)

func getRoutes(repo *cbr.Repository, logger *slog.Logger) *http.ServeMux {
	currencyHandler := handlers.NewCurrencyHandler(
		repo,
		logger.With(slog.String("component", "http")),
	)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/getRates", currencyHandler.GetRates)

	return mux
}
