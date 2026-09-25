package http

import (
	"cbr-worker/internal/cbr/service"
	"cbr-worker/internal/http/handlers/currency"
	"log/slog"

	"github.com/labstack/echo/v5"
)

func registerRoutes(e *echo.Echo, svc *service.Service, logger *slog.Logger) {
	currencyHandler := currency.New(svc, logger)

	e.GET("/v1/rates", currencyHandler.GetRates)
}
