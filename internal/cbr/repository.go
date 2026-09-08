package cbr

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5"
	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Repository struct {
	pool   *pgxpool.Pool
	logger *slog.Logger
}

func NewRepository(pool *pgxpool.Pool, logger *slog.Logger) *Repository {
	return &Repository{pool: pool, logger: logger}
}

const dbTimeout = 10 * time.Second

func (r *Repository) GetLatestExchangeRates(ctx context.Context) ([]*ExchangeRates, error) {
	const sql = `
SELECT rate_date, curr_code, curr_num_code, rate
FROM exchange_rates
WHERE rate_date = (
    SELECT MAX(rate_date)
    FROM exchange_rates
)
ORDER BY curr_code ASC;
`

	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	rows, err := r.pool.Query(ctx, sql)
	if err != nil {
		r.logger.Error("Failed to query latest exchange rates", slog.Any("error", err))

		return nil, fmt.Errorf("failed to query latest exchange rates: %w", err)
	}

	return r.scanExchangeRateRows(rows)
}

func (r *Repository) GetExchangeRates(ctx context.Context, curr []string, date time.Time) ([]*ExchangeRates, error) {
	const sql = `
SELECT rate_date, curr_code, curr_num_code, rate
FROM exchange_rates
WHERE
	($1::text[] IS NULL OR curr_code = ANY($1))
	AND ($2::date IS NULL OR rate_date = $2)
ORDER BY rate_date DESC, curr_code ASC
`

	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	rows, err := r.pool.Query(ctx, sql,
		curr,
		pgtype.Date{
			Time:  date,
			Valid: !date.IsZero(),
		},
	)
	if err != nil {
		r.logger.Error("Failed to query exchange rates",
			slog.Any("curr", curr),
			slog.Time("date", date),
			slog.Any("error", err),
		)

		return nil, fmt.Errorf("failed to query exchange rates: %w", err)
	}

	return r.scanExchangeRateRows(rows)
}

func (r *Repository) scanExchangeRateRows(rows pgx.Rows) ([]*ExchangeRates, error) {
	//type ExchangeRateRaw struct {
	//	Date            pgtype.Date `db:"rate_date"`
	//	CurrencyCode    string      `db:"curr_code"`
	//	CurrencyNumCode int         `db:"curr_num_code"`
	//	Rate            string      `db:"rate"`
	//}
	//
	//raw, err := pgx.CollectRows(rows, pgx.RowToStructByNameLax[ExchangeRateRaw])
	//if err != nil {
	//	r.logger.Error("Failed to query latest exchange rates", slog.Any("error", err))
	//
	//	return nil, fmt.Errorf("failed to query latest exchange rates: %w", err)
	//}

	//if len(raw) == 0 {
	//	return nil, nil
	//}

	if !rows.Next() {
		return nil, nil
	}

	var currDate pgtype.Date
	var currRate ExchangeRate

	err := rows.Scan(&currDate, &currRate.Code, &currRate.NumCode, &currRate.Rate)
	if err != nil {
		return nil, fmt.Errorf("cannot scan exchange rates: %w", err)
	}

	currExchangeRate := &ExchangeRates{
		Date:  Date{Time: currDate.Time},
		Rates: make([]*ExchangeRate, 0, 100),
	}
	currExchangeRate.Rates = append(currExchangeRate.Rates, &currRate)

	rates := make([]*ExchangeRates, 0, 1)

	for rows.Next() {
		var newDate pgtype.Date

		err := rows.Scan(&newDate, &currRate.Code, &currRate.NumCode, &currRate.Rate)
		if err != nil {
			return nil, fmt.Errorf("cannot scan exchange rates: %w", err)
		}

		if !currExchangeRate.Date.Time.Equal(newDate.Time) {
			rates = append(rates, currExchangeRate)

			currExchangeRate = &ExchangeRates{
				Date: Date{Time: newDate.Time},
			}

			currDate = newDate
		}

		currExchangeRate.Rates = append(currExchangeRate.Rates, currRate.Copy())
	}

	rates = append(rates, currExchangeRate)

	return rates, nil
}

func (r *Repository) InsertExchangeRates(ctx context.Context, rates *ExchangeRates) (int, error) {
	const sql = `
INSERT INTO exchange_rates (
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

ON CONFLICT DO NOTHING;
`

	if len(rates.Rates) == 0 {
		return 0, fmt.Errorf("no currencies provided")
	}

	var textCodes = make([]string, 0, len(rates.Rates))
	var numCodes = make([]int, 0, len(rates.Rates))
	var currRates = make([]string, 0, len(rates.Rates))

	for _, curr := range rates.Rates {
		textCodes = append(textCodes, curr.Code)
		numCodes = append(numCodes, curr.NumCode)
		currRates = append(currRates, curr.Rate)
	}

	res, err := r.pool.Exec(ctx, sql,
		rates.Date.Time,
		textCodes, numCodes,
		currRates,
	)
	if err != nil {
		r.logger.Error("Failed to insert exchange rates",
			slog.Any("err", err),
		)

		return 0, fmt.Errorf("failed to insert exchange rates: %w", err)
	}

	return int(res.RowsAffected()), nil
}
