package currency

import (
	"cbr-worker/internal/cbr"
	"fmt"
	"strings"

	"github.com/labstack/echo/v5"
)

type QueryParams struct {
	Currencies []string `query:"currency"`
	FromDate   cbr.Date `query:"fromDate"`
	ToDate     cbr.Date `query:"toDate"`
}

func (p *QueryParams) Parse(c *echo.Context) error {
	if p == nil {
		return fmt.Errorf("QueryParams must not be nil")
	}

	if err := c.Bind(p); err != nil {
		return err
	}

	var currencies []string
	uniqCurrencies := make(map[string]struct{}, 2)
	if len(p.Currencies) > 0 {
		for curr := range strings.SplitSeq(p.Currencies[0], ",") {
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

	p.Currencies = currencies

	return nil
}
