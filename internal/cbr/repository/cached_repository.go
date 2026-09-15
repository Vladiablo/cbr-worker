package repository

import (
	"cbr-worker/internal/cbr"
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"
)

const cacheUpdateTimeout = time.Minute

type CachedRepository struct {
	repo Repository
	cfg  *CachedRepositoryConfig

	cache            atomic.Pointer[map[time.Time]*cbr.ExchangeRates]
	latestRatesCache atomic.Pointer[[]*cbr.ExchangeRates]

	logger *slog.Logger

	updateInterval time.Duration
}

type CachedRepositoryConfig struct {
	CacheUpdateInterval time.Duration
}

func NewCachedRepository(repo Repository, cfg *CachedRepositoryConfig, logger *slog.Logger) *CachedRepository {
	r := &CachedRepository{
		repo:   repo,
		cfg:    cfg,
		logger: logger,
	}

	var updateInterval time.Duration
	if cfg.CacheUpdateInterval > 0 {
		updateInterval = cfg.CacheUpdateInterval
	} else {
		updateInterval = 5 * time.Minute
	}

	r.updateInterval = updateInterval

	return r
}

func (r *CachedRepository) Init(ctx context.Context) error {
	if err := r.updateCache(ctx); err != nil {
		return err
	}

	go func() {
		ticker := time.NewTicker(r.updateInterval)
		defer ticker.Stop()

		for {
			select {
			case <-ctx.Done():
				return
			case <-ticker.C:
				_ = r.updateCache(ctx)
			}
		}
	}()

	return nil
}

func (r *CachedRepository) updateCache(ctx context.Context) error {
	r.logger.Info("Updating CachedRepository cache...")

	ctx, cancel := context.WithTimeout(ctx, cacheUpdateTimeout)
	defer cancel()

	latestRates, err := r.repo.GetLatestExchangeRates(ctx, nil)
	if err != nil {
		r.logger.Error("Failed to update latest exchange rates cache", slog.Any("error", err))

		return err
	}

	if len(latestRates) == 0 {
		r.logger.Error("Failed to update latest exchange rates cache due to empty response")

		return err
	}

	rates, err := r.repo.GetExchangeRatesByDates(ctx, nil, time.Time{}, latestRates[0].Date.Time)
	if err != nil {
		r.logger.Error("Failed to update exchange rates cache", slog.Any("error", err))

		return err
	}

	if len(rates) == 0 {
		r.logger.Error("Failed to update exchange rates cache due to empty response")

		return err
	}

	newRatesCache := make(map[time.Time]*cbr.ExchangeRates, len(rates))
	for _, rate := range rates {
		newRatesCache[rate.Date.Time] = rate
	}

	r.latestRatesCache.Store(&latestRates)
	r.cache.Store(&newRatesCache)

	r.logger.Info("Done updating CachedRepository cache")

	return nil
}

func (r *CachedRepository) GetLatestExchangeRates(_ context.Context, curr []string) ([]*cbr.ExchangeRates, error) {
	latestRatesCache := r.latestRatesCache.Load()
	if latestRatesCache == nil {
		r.logger.Error("Latest rates cache is empty")

		return nil, fmt.Errorf("cache is empty")
	}

	// TODO: implement currency filter
	if len(curr) > 0 {
		result := cbr.ExchangeRates{
			Date:  cbr.Date{},
			Rates: make([]*cbr.ExchangeRate, 0, len(curr)),
		}

		// TODO: Impl actual filtering

		return []*cbr.ExchangeRates{
			&result,
		}, nil
	}

	return *latestRatesCache, nil
}

func (r *CachedRepository) GetExchangeRatesByDates(_ context.Context, curr []string, from, to time.Time) ([]*cbr.ExchangeRates, error) {
	cache := r.cache.Load()
	if cache == nil {
		r.logger.Error("Cache is empty")

		return nil, fmt.Errorf("cache is empty")
	}

	// TODO: implement currency filter
	if len(curr) > 0 {
		//return r.repo.GetExchangeRatesByDates(ctx, curr, from, to)
	}

	result := make([]*cbr.ExchangeRates, 0, int(to.Sub(from).Hours())/24)
	for date := from; !date.After(to); date = date.AddDate(0, 0, 1) {
		rates, ok := (*cache)[date]
		if ok {
			result = append(result, rates)
		}
	}

	if len(result) == 0 {
		return nil, nil
	}

	return result, nil
}

func (r *CachedRepository) InsertExchangeRates(ctx context.Context, rates *cbr.ExchangeRates) (int, error) {
	return r.repo.InsertExchangeRates(ctx, rates)
}
