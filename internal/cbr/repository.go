package cbr

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Currency struct {
	Code    string
	NumCode int
	Rate    string
}

type ExchangeRates struct {
	Date       time.Time
	Currencies []*Currency
}

type Repository struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func NewRepository(pool *pgxpool.Pool) *Repository {
	return &Repository{pool: pool}
}

const dbTimeout = 10 * time.Second

const selectLatestExchangeRatesQuery = `SELECT rate_date, curr_code, curr_num_code, rate
FROM exchange_rates
WHERE rate_date = (
    SELECT MAX(rate_date)
    FROM exchange_rates
);`

func (r *Repository) GetLatestExchangeRates(ctx context.Context) (*ExchangeRates, error) {
	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	rows, err := r.pool.Query(ctx, selectLatestExchangeRatesQuery)
	if err != nil {
		r.logger.Error("Failed to query latest exchange rates", slog.Any("error", err))

		return nil, fmt.Errorf("failed to query latest exchange rates: %w", err)
	}
	defer rows.Close()

	var rates ExchangeRates
	var date pgtype.Date

	for rows.Next() {
		var currency Currency

		err = rows.Scan(&date, &currency.Code, &currency.NumCode, &currency.Rate)
		if err != nil {
			r.logger.Error("Failed to scan exchange rate", slog.Any("error", err))

			return nil, fmt.Errorf("failed to scan exchange rate: %w", err)
		}

		rates.Currencies = append(rates.Currencies, &currency)
	}

	rates.Date = date.Time

	return &rates, nil
}
