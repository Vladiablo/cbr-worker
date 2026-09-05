package internal

import (
	"cbr-worker/internal/cbr"
	"context"
	"fmt"
	"log/slog"
	"strconv"
	"strings"
	"time"

	"github.com/jackc/pgx/v5/pgxpool"
)

type Collector struct {
	cbrClient *cbr.Client
	pool      *pgxpool.Pool
	logger    *slog.Logger
}

func NewCollector(cbrClient *cbr.Client, pool *pgxpool.Pool, logger *slog.Logger) *Collector {
	return &Collector{cbrClient: cbrClient, pool: pool, logger: logger}
}

const insertExchangeRateQuery = `INSERT INTO exchange_rates (
	rate_date,
	curr_code,
	curr_num_code,
	rate
)
SELECT 
	$1::date,
	x.str_code,
	x.num_code,
	x.rate::decimal
FROM unnest(
     $2::text[],
     $3::int[],
     $4::text[]
) AS x(str_code, num_code, rate)

ON CONFLICT DO NOTHING;`

func (c *Collector) Collect(ctx context.Context) error {
	rates, err := c.cbrClient.GetRates(ctx)
	if err != nil {
		return fmt.Errorf("failed to get rates: %w", err)
	}

	date, err := time.Parse("02.01.2006", rates.Date)
	if err != nil {
		return fmt.Errorf("failed to parse date: %w", err)
	}

	var textCodes []string
	var numCodes []int
	var currRates []string

	for _, curr := range rates.Currencies {
		numCode, err := strconv.ParseInt(curr.NumCode, 10, 64)
		if err != nil {
			return fmt.Errorf("failed to parse num code %s: %w", curr.NumCode, err)
		}

		rate := strings.Replace(curr.VunitRate, ",", ".", 1)

		textCodes = append(textCodes, curr.CharCode)
		numCodes = append(numCodes, int(numCode))
		currRates = append(currRates, rate)
	}

	res, err := c.pool.Exec(ctx, insertExchangeRateQuery,
		date,
		textCodes, numCodes,
		currRates,
	)
	if err != nil {
		c.logger.Error("failed to insert exchange rates",
			slog.Any("err", err),
		)

		return fmt.Errorf("failed to insert exchange rates: %w", err)
	}

	c.logger.Info("Successfully inserted exchange rates",
		slog.String("date", date.Format(time.DateOnly)),
		slog.Int64("new", res.RowsAffected()),
	)

	return nil
}
