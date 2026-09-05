package http

import (
	"cbr-worker/internal/http/handlers"
	"log/slog"
	"net/http"

	"github.com/jackc/pgx/v5/pgxpool"
)

func getRoutes(pool *pgxpool.Pool, logger *slog.Logger) *http.ServeMux {
	currencyHandler := handlers.NewCurrencyHandler(
		pool,
		logger.With(slog.String("component", "http")),
	)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/getRates", currencyHandler.GetRates)

	return mux
}
