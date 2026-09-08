package cbr

import (
	"context"
	"fmt"
	"log/slog"
	"time"

	"github.com/jackc/pgx/v5/pgtype"
	"github.com/jackc/pgx/v5/pgxpool"
)

type Date struct {
	time.Time
}

func (d *Date) MarshalJSON() ([]byte, error) {
	return []byte(d.Format("\"2006-01-02\"")), nil
}

type Currency struct {
	Code    string `json:"code"`
	NumCode int    `json:"num_code"`
	Rate    string `json:"rate"`
}

type ExchangeRates struct {
	Date       Date        `json:"date"`
	Currencies []*Currency `json:"currencies"`
}

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
	defer rows.Close()

	var rates ExchangeRates
	var date pgtype.Date

	for rows.Next() {
		var currency Currency

		err = rows.Scan(&date, &currency.Code, &currency.NumCode, &currency.Rate)
		if err != nil {
			r.logger.Error("Failed to scan exchange rate", slog.Any("error", err))

			return nil, fmt.Errorf("failed to scan exchange rate: %w", err)
		}

		rates.Currencies = append(rates.Currencies, &currency)
	}

	rates.Date.Time = date.Time

	return []*ExchangeRates{&rates}, nil
}

func (r *Repository) GetExchangeRates(ctx context.Context, curr string, date time.Time) ([]*ExchangeRates, error) {
	const sql = `
SELECT rate_date, curr_code, curr_num_code, rate
FROM exchange_rates
WHERE ($1::text IS NULL OR curr_code = $1)
AND ($2::date IS NULL OR rate_date = $2)
ORDER BY rate_date DESC, curr_code ASC;
`

	ctx, cancel := context.WithTimeout(ctx, dbTimeout)
	defer cancel()

	rows, err := r.pool.Query(ctx, sql,
		pgtype.Text{
			String: curr,
			Valid:  curr != "",
		},
		pgtype.Date{
			Time:  date,
			Valid: !date.IsZero(),
		},
	)
	if err != nil {
		r.logger.Error("Failed to query exchange rates",
			slog.String("curr", curr),
			slog.Time("date", date),
			slog.Any("error", err),
		)

		return nil, fmt.Errorf("failed to query exchange rates: %w", err)
	}
	defer rows.Close()

	result := make([]*ExchangeRates, 0, 1)

	{
		var rates *ExchangeRates
		var prevDate, date pgtype.Date

		for rows.Next() {
			var currency Currency

			prevDate = date

			err = rows.Scan(&date, &currency.Code, &currency.NumCode, &currency.Rate)
			if err != nil {
				r.logger.Error("Failed to scan exchange rate", slog.Any("error", err))

				return nil, fmt.Errorf("failed to scan exchange rate: %w", err)
			}

			if date != prevDate {
				rates = &ExchangeRates{}
				rates.Date.Time = date.Time

				result = append(result, rates)
			}

			if rates == nil {
				return nil, fmt.Errorf("rates was nil")
			}

			rates.Currencies = append(rates.Currencies, &currency)
		}
	}

	return result, nil
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

	if len(rates.Currencies) == 0 {
		return 0, fmt.Errorf("no currencies provided")
	}

	var textCodes = make([]string, 0, len(rates.Currencies))
	var numCodes = make([]int, 0, len(rates.Currencies))
	var currRates = make([]string, 0, len(rates.Currencies))

	for _, curr := range rates.Currencies {
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
