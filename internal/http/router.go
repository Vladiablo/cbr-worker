package http

import (
	"cbr-worker/internal/cbr/service"
	"cbr-worker/internal/http/handlers/currency"
	"log/slog"
	"net/http"
)

func getRoutes(svc *service.Service, logger *slog.Logger) *http.ServeMux {
	currencyHandler := currency.New(
		svc,
		logger.With(slog.String("component", "http")),
	)

	mux := http.NewServeMux()
	mux.HandleFunc("GET /v1/rates", currencyHandler.GetRates)

	return mux
}
