package currency

import (
	"fmt"
	"net/url"
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
	uniqCurrencies := make(map[string]struct{}, 2)
	rawCurrencies := query.Get("currency")
	if len(rawCurrencies) > 0 {
		for curr := range strings.SplitSeq(rawCurrencies, ",") {
			curr := strings.TrimSpace(curr)
			if len(curr) == 0 {
				continue
			}

			if _, ok := uniqCurrencies[curr]; ok {
				continue
			}

			uniqCurrencies[curr] = struct{}{}
			currencies = append(currencies, curr)
		}
	}

	p.fromDate = fromDate
	p.toDate = toDate
	p.currencies = currencies

	return nil
}
