package http

import (
	"github.com/labstack/echo/v5"
)

type Handlers struct {
	GetRates echo.HandlerFunc
}

func RegisterRoutes(srv *Server, handlers *Handlers) {
	srv.echo.GET("/v1/rates", handlers.GetRates)
}
