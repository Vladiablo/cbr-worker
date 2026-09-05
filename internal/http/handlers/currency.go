package handlers

import (
	"cbr-worker/internal/cbr"
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"
)

type CurrencyHandler struct {
	repo   *cbr.Repository
	logger *slog.Logger
}

type Currency struct {
	Code    string `json:"curr_code"`
	NumCode int    `json:"curr_num_code"`
	Rate    string `json:"rate"`
}

type RatesResponse struct {
	Date       string      `json:"date"`
	Currencies []*Currency `json:"currencies"`
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

const repoTimeout = 10 * time.Second

func (h *CurrencyHandler) GetRates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(r.Context(), repoTimeout)
	defer cancel()

	rates, err := h.repo.GetLatestExchangeRates(ctx)
	if err != nil {
		h.logger.Error("Failed to get latest exchange rates", slog.Any("error", err))
		h.writeErr(w, err, http.StatusInternalServerError)
	}

	var ratesResponse RatesResponse
	ratesResponse.Date = rates.Date.Format("02.01.2006")
	for _, rate := range rates.Currencies {
		ratesResponse.Currencies = append(ratesResponse.Currencies, &Currency{
			Code:    rate.Code,
			NumCode: rate.NumCode,
			Rate:    rate.Rate,
		})
	}

	err = json.NewEncoder(w).Encode(ratesResponse)
	if err != nil {
		h.writeErr(w, err, http.StatusInternalServerError)
		return
	}
	h.logger.Info("Done")
}
