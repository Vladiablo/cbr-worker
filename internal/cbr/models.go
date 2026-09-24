package cbr

import "time"

type Date struct {
	time.Time
}

func (d *Date) MarshalJSON() ([]byte, error) {
	return []byte(d.Format("\"2006-01-02\"")), nil
}

// ExchangeRate model info
//
//	@Description	Exchange rate for given currency
type ExchangeRate struct {
	Code    string `json:"code" example:"USD"`           // Currency text code according ISO 4217
	NumCode int    `json:"num_code" example:"840"`       // Currency numeric code according to ISO 4217 (without leading zeros
	Rate    string `json:"rate" example:"28.4821000000"` // Currency exchange rate
}

func (r *ExchangeRate) Copy() *ExchangeRate {
	return new(*r)
}

// ExchangeRates model info
//
//	@Description	Array of exchange rates for given currencies at given date
type ExchangeRates struct {
	Date  Date            `json:"date" swaggertype:"primitive,string" example:"2006-01-11"` // Date of exchange rates
	Rates []*ExchangeRate `json:"rates"`                                                    // Array of exchange rates
}
