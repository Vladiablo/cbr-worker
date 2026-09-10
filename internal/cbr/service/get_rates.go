package service

import (
	"cbr-worker/internal/cbr"
	"context"
	"fmt"
	"time"
)

func (svc *Service) GetRates(ctx context.Context, params *GetRatesParams) ([]*cbr.ExchangeRates, error) {
	if err := params.Validate(); err != nil {
		return nil, GetRatesParamsError{err: err}
	}

	var rates []*cbr.ExchangeRates
	var err error

	if params.haveDates() {
		rates, err = svc.repo.GetExchangeRatesByDates(ctx, params.Currencies, params.FromDate, params.ToDate)
	} else {
		rates, err = svc.repo.GetLatestExchangeRates(ctx, params.Currencies)
	}

	if err != nil {
		return nil, fmt.Errorf("failed to get rates: %w", err)
	}

	return rates, nil
}

type GetRatesParamsError struct {
	err error
}

func (e GetRatesParamsError) Error() string {
	return fmt.Sprintf("failed to validate parameters: %v", e.err.Error())
}

type GetRatesParams struct {
	Currencies []string
	FromDate   time.Time
	ToDate     time.Time
}

const maxFromAndToDateDistance = 30 * 24 * time.Hour

func (p *GetRatesParams) Validate() error {
	if p.ToDate.IsZero() != p.FromDate.IsZero() {
		return fmt.Errorf("both fromDate and toDate must be specified")
	}

	if p.ToDate.Before(p.FromDate) {
		return fmt.Errorf("toDate must be after fromDate")
	}

	if p.ToDate.Sub(p.FromDate) > maxFromAndToDateDistance {
		return fmt.Errorf("fromDate and toDate distance must not exceed %v", maxFromAndToDateDistance)
	}

	return nil
}

func (p *GetRatesParams) haveDates() bool {
	return !p.FromDate.IsZero() && !p.ToDate.IsZero()
}
