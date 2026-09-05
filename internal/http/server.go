package http

import (
	"cbr-worker/internal/cbr"
	"context"
	"fmt"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
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

	err := s.srv.ListenAndServe()
	if err != nil {
		s.logger.Error("Failed to start HTTP server", slog.Any("error", err))

		return fmt.Errorf("failed to start HTTP server: %w", err)
	}

	return nil
}
