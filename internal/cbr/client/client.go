package client

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"github.com/cenkalti/backoff/v7"
	"resty.dev/v3"
)

type Client struct {
	httpClient *resty.Client
	cfg        *Config

	logger *slog.Logger
}

type RawCurrency struct {
	ID        string `xml:"ID,attr"`
	NumCode   string
	CharCode  string
	Nominal   uint
	Name      string
	Value     string
	VunitRate string
}

type RatesResponse struct {
	Date       string         `xml:"Date,attr"`
	Name       string         `xml:"name,attr"`
	Currencies []*RawCurrency `xml:"Valute"`
}

func identicalCharsetReader(_ string, input io.Reader) (io.Reader, error) {
	return input, nil
}

func New(httpClient *http.Client, cfg *Config, logger *slog.Logger) (*Client, error) {
	if err := cfg.Validate(); err != nil {
		return nil, fmt.Errorf("config validation failed: %w", err)
	}

	client := resty.NewWithClient(httpClient)
	client.SetBaseURL(cfg.BaseUrl)

	// client.SetDebug(true)

	return &Client{logger: logger, cfg: cfg, httpClient: client}, nil
}

func (c *Client) executeGetRates(ctx context.Context, date time.Time) (*RatesResponse, error) {
	const unexpectedStatusCodeResponseBodySizeLimit = 4 * 1024

	req := c.httpClient.R().WithContext(ctx)

	if !date.IsZero() {
		req = req.SetQueryParam("date_req", date.Format("02/01/2006"))
	}

	resp, err := req.Get("/scripts/XML_daily_eng.asp")
	if err != nil {
		c.logger.Error("Failed to execute GET rates request", "error", err)

		return nil, fmt.Errorf("failed to execute request to CBR: %w", err)
	}
	defer func() { _ = resp.Body.Close() }()

	if resp.StatusCode() != http.StatusOK {
		body := make([]byte, unexpectedStatusCodeResponseBodySizeLimit)

		reader := io.LimitReader(resp.Body, unexpectedStatusCodeResponseBodySizeLimit)
		n, err := reader.Read(body)
		if n == 0 || err != nil {
			c.logger.Warn("Failed to get response body during processing unexpected HTTP status code",
				slog.Int("n", n),
				slog.Any("err", err),
			)
		}
		c.logger.Error("Received unexpected HTTP status code from CBR",
			slog.Int("statusCode", resp.StatusCode()),
			slog.String("body", string(body[:n])),
		)

		err = fmt.Errorf("CBR responded with unexpected HTTP status code: %d", resp.StatusCode())
		if resp.StatusCode() >= 400 && resp.StatusCode() < 500 {
			return nil, backoff.Permanent(err)
		}

		return nil, err
	}

	// Dangerous conversion from Windows-1251 to UTF-8
	decoder := xml.NewDecoder(resp.Body)
	decoder.CharsetReader = identicalCharsetReader

	var result RatesResponse
	err = decoder.Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to decode CBR response body: %w", err)
	}

	return &result, nil
}

func (c *Client) GetRates(ctx context.Context, date time.Time) (*RatesResponse, error) {
	result, err := backoff.Retry(ctx,
		func() (*RatesResponse, error) {
			return c.executeGetRates(ctx, date)
		},
		backoff.WithMaxTries(uint(c.cfg.MaxRetries+1)),
		backoff.WithBackOff(&backoff.ExponentialBackOff{
			InitialInterval:     c.cfg.RetryInitialInterval,
			RandomizationFactor: c.cfg.RetryRandomizationFactor,
			Multiplier:          c.cfg.RetryMultiplier,
			MaxInterval:         c.cfg.RetryMaxInterval,
		}),
	)

	if err != nil {
		return nil, err
	}

	if len(result.Currencies) == 0 {
		return nil, fmt.Errorf("no currencies found")
	}

	return result, nil
}
