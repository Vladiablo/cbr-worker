package cbr

import "time"

type Date struct {
	time.Time
}

func (d *Date) MarshalJSON() ([]byte, error) {
	return []byte(d.Format("\"2006-01-02\"")), nil
}

type ExchangeRate struct {
	Code    string `json:"code"`
	NumCode int    `json:"num_code"`
	Rate    string `json:"rate"`
}

func (r *ExchangeRate) Copy() *ExchangeRate {
	return new(*r)
}

type ExchangeRates struct {
	Date  Date            `json:"date"`
	Rates []*ExchangeRate `json:"rates"`
}
