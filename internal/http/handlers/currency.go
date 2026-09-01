package handlers

import (
	"cbr-worker/internal/cbr"
	"encoding/json"
	"log/slog"
	"net/http"
)

type CurrencyHandler struct {
	cbrClient *cbr.Client
	logger    *slog.Logger
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewCurrencyHandler(cbrClient *cbr.Client, logger *slog.Logger) *CurrencyHandler {
	return &CurrencyHandler{cbrClient: cbrClient, logger: logger}
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

	cbrResp, err := h.cbrClient.GetRates(r.Context())
	if err != nil {
		h.writeErr(w, err, http.StatusInternalServerError)
		return
	}

	err = json.NewEncoder(w).Encode(cbrResp)
	if err != nil {
		h.writeErr(w, err, http.StatusInternalServerError)
		return
	}
	h.logger.Info("Done")
}
