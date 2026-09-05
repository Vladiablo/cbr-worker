package handlers

import (
	"context"
	"encoding/json"
	"log/slog"
	"net/http"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type CurrencyHandler struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

type Currency struct {
	Code    string `json:"curr_code"`
	NumCode int    `json:"curr_num_code"`
	Rate    string `json:"rate"`
}

type RatesResponse struct {
	Date       string     `json:"date"`
	Currencies []Currency `json:"currencies"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

func NewCurrencyHandler(pool *pgxpool.Pool, logger *slog.Logger) *CurrencyHandler {
	return &CurrencyHandler{pool: pool, logger: logger}
}

func (h *CurrencyHandler) writeErr(w http.ResponseWriter, err error, statusCode int) {
	w.WriteHeader(statusCode)
	err = json.NewEncoder(w).Encode(ErrorResponse{err.Error()})
	if err != nil {
		h.logger.Error("Failed to encode error response", slog.Any("error", err))
	}
}

const dbTimeout = 10 * time.Second

const selectLatestExchangeRatesQuery = `SELECT rate_date, curr_code, curr_num_code, rate
FROM exchange_rates
WHERE rate_date = (
    SELECT MAX(rate_date)
    FROM exchange_rates
);`

func (h *CurrencyHandler) GetRates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	ctx, cancel := context.WithTimeout(r.Context(), dbTimeout)
	defer cancel()

	rows, err := h.pool.Query(ctx, selectLatestExchangeRatesQuery)
	if err != nil {
		h.logger.Error("Failed to query latest exchange rates", slog.Any("error", err))
		h.writeErr(w, err, http.StatusInternalServerError)
		return
	}
	defer rows.Close()

	var ratesResponse RatesResponse
	var date pgtype.Date

	for rows.Next() {
		var currency Currency

		err = rows.Scan(&date, &currency.Code, &currency.NumCode, &currency.Rate)
		if err != nil {
			h.logger.Error("Failed to scan exchange rate", slog.Any("error", err))
			h.writeErr(w, err, http.StatusInternalServerError)
			return
		}

		ratesResponse.Currencies = append(ratesResponse.Currencies, currency)
	}

	ratesResponse.Date = date.Time.Format(time.DateOnly)

	err = json.NewEncoder(w).Encode(ratesResponse)
	if err != nil {
		h.writeErr(w, err, http.StatusInternalServerError)
		return
	}
	h.logger.Info("Done")
}
