package handlers

import (
	"cbr-worker/internal/cbr"
	"encoding/json"
	"fmt"
	"log/slog"
	"net/http"
	"time"
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

	var rates []*cbr.ExchangeRates
	var err error

	var date time.Time
	var currency string

	if query := r.URL.Query(); len(query) > 0 {
		strDate := query.Get("date")
		if strDate != "" {
			date, err = time.Parse(time.DateOnly, strDate)
			if err != nil {
				h.logger.Error("Failed to parse date", slog.Any("date", strDate))
				h.writeErr(w, fmt.Errorf("invalid date format"), http.StatusBadRequest)

				return
			}
		}

		currency = query.Get("currency")
	}

	if currency != "" || !date.IsZero() {
		rates, err = h.repo.GetExchangeRates(r.Context(), currency, date)
	} else {
		rates, err = h.repo.GetLatestExchangeRates(r.Context())
	}

	if err != nil {
		h.logger.Error("Failed to get exchange rates", slog.Any("error", err))
		h.writeErr(w, err, http.StatusInternalServerError)

		return
	}

	if len(rates) == 0 {
		h.writeErr(w, fmt.Errorf("no exchange rates found"), http.StatusNotFound)

		return
	}

	err = json.NewEncoder(w).Encode(rates)
	if err != nil {
		h.writeErr(w, err, http.StatusInternalServerError)
	}
}
