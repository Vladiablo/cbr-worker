package repository

import (
	"cbr-worker/internal/cbr"
	"context"
	"time"
)

type Repository interface {
	GetLatestExchangeRates(ctx context.Context, curr []string) ([]*cbr.ExchangeRates, error)
	GetExchangeRatesByDates(ctx context.Context, curr []string, from, to time.Time) ([]*cbr.ExchangeRates, error)

	InsertExchangeRates(ctx context.Context, rates *cbr.ExchangeRates) (int, error)
}
