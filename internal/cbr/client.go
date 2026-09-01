package cbr

import (
	"cbr-worker/internal/http/handlers"
	"log/slog"
	"net/http"
)

type CbrClient struct {
	Logger *slog.Logger
}

func (cbr *CbrClient) HandleGetCurs(w http.ResponseWriter, r *http.Request) {
	handlers.HandleGetCurs(w, r, cbr.Logger)
}
