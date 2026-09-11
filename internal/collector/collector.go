package collector

import (
	"cbr-worker/internal/cbr"
	"context"
	"fmt"
	"iter"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"golang.org/x/sync/errgroup"
)

type Collector struct {
	cbrClient *cbr.Client
	repo      *cbr.Repository
	cfg       *Config

	logger *slog.Logger
}

func New(cbrClient *cbr.Client, repo *cbr.Repository, cfg *Config, logger *slog.Logger) *Collector {
	return &Collector{cbrClient: cbrClient, repo: repo, cfg: cfg, logger: logger}
}

func generateDateRange(from, to time.Time) iter.Seq[time.Time] {
	return func(yield func(time.Time) bool) {
		for !from.After(to) {
			if !yield(from) {
				return
			}
			from = from.AddDate(0, 0, 1)
		}
	}
}

func (c *Collector) Collect(ctx context.Context) error {
	if err := c.cfg.Validate(); err != nil {
		return fmt.Errorf("config validation failed: %w", err)
	}

	// TODO: more reliable method to get latest rates date
	if !c.cfg.FromDate.IsZero() && c.cfg.ToDate.IsZero() {
		c.cfg.ToDate = time.Now().Truncate(24 * time.Hour)
	}

	eg, ctx := errgroup.WithContext(ctx)

	dates := make(chan time.Time)
	eg.Go(func() error {
		defer close(dates)

		for date := range generateDateRange(c.cfg.FromDate, c.cfg.ToDate) {
			select {
			case dates <- date:
			case <-ctx.Done():
				c.logger.Info("Collector shutting down...")

				return nil
			}
		}

		return nil
	})

	var workersCnt int
	if c.cfg.Concurrency == 0 {
		workersCnt = 1
	} else {
		workersCnt = c.cfg.Concurrency
	}

	for range workersCnt {
		eg.Go(func() error {
			for date := range dates {
				if err := c.CollectDate(context.Background(), date); err != nil {
					c.logger.Error("Failed to collect date", "date", date, "err", err)

					return fmt.Errorf("failed to collect date %s: %w", date.Format(time.DateOnly), err)
				}
			}

			return nil
		})
	}

	if err := eg.Wait(); err != nil {
		return err
	}

	return nil
}

const defaultTimeout = 30 * time.Second

func (c *Collector) CollectDate(ctx context.Context, wantDate time.Time) error {
	var timeout time.Duration
	if c.cfg.Timeout == 0 {
		timeout = defaultTimeout
	} else {
		timeout = c.cfg.Timeout
	}

	ctx, cancel := context.WithTimeout(ctx, timeout)
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
