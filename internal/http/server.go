package http

import (
	"cbr-worker/internal/cbr"
	"context"
	"errors"
	"fmt"
	"log/slog"
	"net"
	"net/http"
	"time"
)

type Server struct {
	srv *http.Server

	logger *slog.Logger
}

func NewServer(addr string, repo *cbr.Repository, logger *slog.Logger) *Server {
	srv := &http.Server{
		Addr:         addr,
		Handler:      getRoutes(repo, logger),
		ReadTimeout:  30 * time.Second,
		WriteTimeout: 30 * time.Second,
	}

	return &Server{
		srv: srv,

		logger: logger,
	}
}

func (s *Server) Start(ctx context.Context) error {
	s.logger.Info("Starting HTTP server...", slog.String("addr", s.srv.Addr))

	// Note: lc.Listen doesn't respect cancelled context, so we do our own check
	if ctx.Err() != nil {
		return ctx.Err()
	}

	var lc net.ListenConfig
	listener, err := lc.Listen(ctx, "tcp", s.srv.Addr)
	if err != nil {
		return fmt.Errorf("http server failed to listen: %w", err)
	}

	s.logger.Info("HTTP server started", slog.String("addr", s.srv.Addr))

	go func() {
		<-ctx.Done()

		_ = s.shutdown(context.Background())
	}()

	if err := s.srv.Serve(listener); err != nil {
		if errors.Is(err, http.ErrServerClosed) {
			s.logger.Info("HTTP server stopped")

			return nil
		}

		s.logger.Error("Failed to start HTTP server", slog.Any("error", err))

		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	return nil
}

func (s *Server) shutdown(ctx context.Context) error {
	s.logger.Info("Shutting down HTTP server...")

	return s.srv.Shutdown(ctx)
}
