package handlers

import (
	"cbr-worker/internal/cbr"
	"encoding/json"
	"log/slog"
	"net/http"
)

type CurrencyHandler struct {
	repo   *cbr.Repository
	logger *slog.Logger
}
type ErrorResponse struct {
	Error string `json:"error"`
}

func NewCurrencyHandler(repo *cbr.Repository, logger *slog.Logger) *CurrencyHandler {
	return &CurrencyHandler{repo: repo, logger: logger}
}

func (h *CurrencyHandler) writeErr(w http.ResponseWriter, err error, statusCode int) {
	w.WriteHeader(statusCode)
	err = json.NewEncoder(w).Encode(ErrorResponse{err.Error()})
	if err != nil {
		h.logger.Error("Failed to encode error response", slog.Any("error", err))
	}
}

func (h *CurrencyHandler) GetRates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	rates, err := h.repo.GetLatestExchangeRates(r.Context())
	if err != nil {
		h.logger.Error("Failed to get latest exchange rates", slog.Any("error", err))
		h.writeErr(w, err, http.StatusInternalServerError)
	}

	err = json.NewEncoder(w).Encode(rates)
	if err != nil {
		h.writeErr(w, err, http.StatusInternalServerError)
		return
	}
	h.logger.Info("Done")
}
