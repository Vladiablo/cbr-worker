package http

import (
	"cbr-worker/internal/http/handlers/currency"
)

func RegisterRoutes(srv *Server, currencyHandler *currency.Handler) {
	v1 := srv.echo.Group("/v1")

	currency.RegisterRoutes(v1, currencyHandler)
}
