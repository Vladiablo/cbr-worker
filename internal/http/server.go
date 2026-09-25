package http

import (
	"cbr-worker/internal/cbr/service"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"

	"github.com/labstack/echo/v5"
	"github.com/labstack/echo/v5/middleware"
)

type Server struct {
	srv  *echo.Echo
	addr string

	logger *slog.Logger
}

func NewServer(addr string, svc *service.Service, logger *slog.Logger) *Server {
	e := echo.NewWithConfig(echo.Config{
		Logger:           logger,
		HTTPErrorHandler: httpErrorHandler,
	})

	e.Use(middleware.Recover())
	e.Use(middleware.ContextTimeout(30 * time.Second))

	registerRoutes(e, svc, logger.With(slog.String("component", "http")))

	return &Server{
		srv:  e,
		addr: addr,

		logger: logger,
	}
}

func (s *Server) Start(ctx context.Context, gracefulTimeout time.Duration) error {
	s.logger.Info("Starting HTTP server...", slog.String("addr", s.addr))

	// Note: lc.Listen doesn't respect cancelled context, so we do our own check
	if ctx.Err() != nil {
		return ctx.Err()
	}

	var lc net.ListenConfig
	listener, err := lc.Listen(ctx, "tcp", s.addr)
	if err != nil {
		return fmt.Errorf("http server failed to listen: %w", err)
	}

	go func() {
		<-ctx.Done()

		s.logger.Info("Shutting down HTTP server...")
	}()

	sc := echo.StartConfig{
		Address:         s.addr,
		HideBanner:      true,
		Listener:        listener,
		GracefulTimeout: gracefulTimeout,
	}

	if err := sc.Start(ctx, s.srv); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			s.logger.Info("HTTP server stopped")

			return nil
		}

		s.logger.Error("Failed to start HTTP server", slog.Any("error", err))

		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	return nil
}
