package currency

import (
	"cbr-worker/internal/cbr"
	"strings"
)

type GetRatesRequest struct {
	Currencies string   `query:"currency"`
	FromDate   cbr.Date `query:"fromDate"`
	ToDate     cbr.Date `query:"toDate"`
}

func (r *GetRatesRequest) GetUniqCurrencies() []string {
	if len(r.Currencies) == 0 {
		return nil
	}

	var currencies []string
	uniqCurrencies := make(map[string]struct{}, 2)

	for curr := range strings.SplitSeq(r.Currencies, ",") {
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

	return currencies
}
