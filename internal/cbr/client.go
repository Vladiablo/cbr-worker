package cbr

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"

	"resty.dev/v3"
)

type Client struct {
	httpClient *resty.Client

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

const getRatesRequestTImeout = time.Second * 5
const unexpectedStatusCodeResponseBodySizeLimit = 4 * 1024

func identicalCharsetReader(_ string, input io.Reader) (io.Reader, error) {
	return input, nil
}

func NewClient(httpClient *http.Client, logger *slog.Logger) *Client {
	client := resty.NewWithClient(httpClient)
	client.SetBaseURL("http://www.cbr.ru")
	//client.SetDebug(true)

	return &Client{logger: logger, httpClient: client}
}

func (c *Client) GetRates(ctx context.Context, date time.Time) (*RatesResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, getRatesRequestTImeout)
	defer cancel()

	req := c.httpClient.R().WithContext(ctx)

	if !date.IsZero() {
		req = req.SetQueryParam("date_req", date.Format("02/01/2006"))
	}

	resp, err := req.Get("/scripts/XML_daily_eng.asp")
	if err != nil {
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

		return nil, fmt.Errorf("CBR responded with unexpected HTTP status code: %d", resp.StatusCode())
	}

	// Dangerous conversion from Windows-1251 to UTF-8
	decoder := xml.NewDecoder(resp.Body)
	decoder.CharsetReader = identicalCharsetReader

	var result RatesResponse
	err = decoder.Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to decode CBR response body: %w", err)
	}

	if len(result.Currencies) == 0 {
		return nil, fmt.Errorf("no currencies found")
	}

	return &result, nil
}
