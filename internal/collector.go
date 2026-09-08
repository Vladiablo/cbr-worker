package internal

import (
	"cbr-worker/internal/cbr"
	"context"
	"fmt"
	"iter"
	"log/slog"
	"strconv"
	"strings"
	"time"
)

type Collector struct {
	cbrClient *cbr.Client
	repo      *cbr.Repository
	logger    *slog.Logger
}

func NewCollector(cbrClient *cbr.Client, repo *cbr.Repository, logger *slog.Logger) *Collector {
	return &Collector{cbrClient: cbrClient, repo: repo, logger: logger}
}

func generateDateRange(from, to time.Time) iter.Seq[time.Time] {
	return func(yield func(time.Time) bool) {
		for !from.After(to) {
			yield(from)
			from = from.AddDate(0, 0, 1)
		}
	}
}

func (c *Collector) Collect(ctx context.Context, from time.Time, to time.Time) error {
	for date := range generateDateRange(from, to) {
		if err := c.CollectDate(ctx, date); err != nil {
			return fmt.Errorf("failed collect date %s: %w", date.Format(time.DateOnly), err)
		}
	}

	return nil
}

func (c *Collector) CollectDate(ctx context.Context, wantDate time.Time) error {
	// TODO: Make timeout configurable
	ctx, cancel := context.WithTimeout(ctx, 30*time.Second)
	defer cancel()

	c.logger.Info("CollectDate",
		slog.String("wantDate", wantDate.Format(time.DateOnly)),
	)

	cbrRates, err := c.cbrClient.GetRates(ctx, wantDate)
	if err != nil {
		return fmt.Errorf("failed to get rates: %w", err)
	}

	date, err := time.Parse("02.01.2006", cbrRates.Date)
	if err != nil {
		return fmt.Errorf("failed to parse date: %w", err)
	}

	rates := cbr.ExchangeRates{
		Date:  cbr.Date{Time: date},
		Rates: make([]*cbr.ExchangeRate, 0, len(cbrRates.Currencies)),
	}

	for i := range cbrRates.Currencies {
		curr := cbrRates.Currencies[i]

		numCode, err := strconv.ParseInt(curr.NumCode, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse num code %s: %w", curr.NumCode, err)
		}

		rate := strings.Replace(curr.VunitRate, ",", ".", 1)

		rates.Rates = append(rates.Rates, &cbr.ExchangeRate{
			Code:    curr.CharCode,
			NumCode: int(numCode),
			Rate:    rate,
		})
	}

	n, err := c.repo.InsertExchangeRates(ctx, &rates)
	if err != nil {
		c.logger.Error("Failed to insert exchange rates",
			slog.Any("err", err),
		)

		return fmt.Errorf("failed to insert exchange rates: %w", err)
	}

	c.logger.Info("Successfully inserted exchange rates",
		slog.String("date", date.Format(time.DateOnly)),
		slog.Int("new", n),
	)

	return nil
}

// TODO: console args: -from date, -to date
// if only from set to to now
// if only to return error
