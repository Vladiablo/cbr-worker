package repository

import (
	"cbr-worker/internal/cbr"
	"context"
	"fmt"
	"log/slog"
	"sync/atomic"
	"time"
)

type CachedRepository struct {
	repo Repository
	cfg  *CachedRepositoryConfig

	cache            atomic.Pointer[map[time.Time]*cbr.ExchangeRates]
	latestRatesCache atomic.Pointer[[]*cbr.ExchangeRates]

	logger *slog.Logger
}

type CachedRepositoryConfig struct {
	CacheUpdateInterval time.Duration
}

func NewCachedRepository(repo Repository, cfg *CachedRepositoryConfig, logger *slog.Logger) *CachedRepository {
	r := &CachedRepository{repo: repo, cfg: cfg, logger: logger}

	var updateInterval time.Duration
	if cfg.CacheUpdateInterval > 0 {
		updateInterval = cfg.CacheUpdateInterval
	} else {
		updateInterval = 5 * time.Minute
	}

	r.updateCache()
	go func() {
		for range time.Tick(updateInterval) {
			r.updateCache()
		}
	}()

	return r
}

const cacheUpdateTimeout = time.Minute

func (r *CachedRepository) updateCache() {
	r.logger.Info("Updating CachedRepository cache...")

	ctx, cancel := context.WithTimeout(context.Background(), cacheUpdateTimeout)
	defer cancel()

	latestRates, err := r.repo.GetLatestExchangeRates(ctx, nil)
	if err != nil {
		r.logger.Error("Failed to update latest exchange rates cache", slog.Any("error", err))

		return
	}

	if len(latestRates) == 0 {
		r.logger.Error("Failed to update latest exchange rates cache due to empty response")

		return
	}

	rates, err := r.repo.GetExchangeRatesByDates(ctx, nil, time.Time{}, latestRates[0].Date.Time)
	if err != nil {
		r.logger.Error("Failed to update exchange rates cache", slog.Any("error", err))

		return
	}

	if len(rates) == 0 {
		r.logger.Error("Failed to update exchange rates cache due to empty response")

		return
	}

	newRatesCache := make(map[time.Time]*cbr.ExchangeRates, len(rates))
	for _, rate := range rates {
		newRatesCache[rate.Date.Time] = rate
	}

	r.latestRatesCache.Store(&latestRates)
	r.cache.Store(&newRatesCache)

	r.logger.Info("Done updating CachedRepository cache")
}

func (r *CachedRepository) GetLatestExchangeRates(ctx context.Context, curr []string) ([]*cbr.ExchangeRates, error) {
	latestRatesCache := r.latestRatesCache.Load()
	if latestRatesCache == nil {
		r.logger.Error("Latest rates cache is empty")

		return nil, fmt.Errorf("cache is empty")
	}

	// TODO: implement currency filter
	if curr != nil {
		return r.repo.GetLatestExchangeRates(ctx, curr)
	}

	return *latestRatesCache, nil
}

func (r *CachedRepository) GetExchangeRatesByDates(ctx context.Context, curr []string, from, to time.Time) ([]*cbr.ExchangeRates, error) {
	cache := r.cache.Load()
	if cache == nil {
		r.logger.Error("Cache is empty")

		return nil, fmt.Errorf("cache is empty")
	}

	// TODO: implement currency filter
	if curr != nil {
		return r.repo.GetExchangeRatesByDates(ctx, curr, from, to)
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
