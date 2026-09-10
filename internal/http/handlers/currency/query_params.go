package currency

import (
	"fmt"
	"net/url"
	"slices"
	"strings"
	"time"
)

type QueryParams struct {
	currencies []string
	fromDate   time.Time
	toDate     time.Time
}

func parseDate(query url.Values, name string) (time.Time, error) {
	strDate := query.Get(name)
	if strDate != "" {
		result, err := time.Parse(time.DateOnly, strDate)
		if err != nil {
			return time.Time{}, fmt.Errorf("%s has invalid format: %w", name, err)
		}

		return result, nil
	}

	return time.Time{}, nil
}

func (p *QueryParams) Parse(query url.Values) error {
	if p == nil {
		return fmt.Errorf("QueryParams must not be nil")
	}

	if len(query) == 0 {
		return nil
	}

	fromDate, err := parseDate(query, "fromDate")
	if err != nil {
		return err
	}

	toDate, err := parseDate(query, "toDate")
	if err != nil {
		return err
	}

	var currencies []string
	rawCurrencies := query.Get("currency")
	if len(rawCurrencies) > 0 {
		currencies = strings.Split(
			rawCurrencies,
			",",
		)
		currencies = slices.DeleteFunc(currencies, func(s string) bool {
			return len(s) == 0
		})
	}

	p.fromDate = fromDate
	p.toDate = toDate
	p.currencies = currencies

	return nil
}
