package cbr

import (
	"context"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
	"time"
)

type Client struct {
	logger     *slog.Logger
	httpClient *http.Client
}

type Currency struct {
	ID        string `xml:"ID,attr" json:"id"`
	NumCode   string `json:"numCode"`
	CharCode  string `json:"charCode"`
	Nominal   uint   `json:"nominal"`
	Name      string `json:"name"`
	Value     string `json:"value"`
	VunitRate string `json:"vunitRate"`
}

type RatesResponse struct {
	Date       string     `xml:"Date,attr" json:"date"`
	Name       string     `xml:"name,attr" json:"name"`
	Currencies []Currency `xml:"Valute" json:"currencies"`
}

const getRatesCbrMirrorUrl = "https://www.cbr-xml-daily.ru/daily_eng_utf8.xml"

const getRatesRequestTImeout = time.Second * 5
const unexpectedStatusCodeResponseBodySizeLimit = 4 * 1024

func indenticalCharsetReader(encoding string, input io.Reader) (io.Reader, error) {
	return input, nil
}

func NewClient(logger *slog.Logger, httpClient *http.Client) *Client {
	return &Client{logger, httpClient}
}

func (c *Client) GetRates(ctx context.Context) (*RatesResponse, error) {
	ctx, cancel := context.WithTimeout(ctx, getRatesRequestTImeout)
	defer cancel()

	req, err := http.NewRequestWithContext(ctx, http.MethodGet, getRatesCbrMirrorUrl, http.NoBody)
	if err != nil {
		return nil, fmt.Errorf("failed to create request to CBR: %w", err)
	}

	resp, err := c.httpClient.Do(req)
	if err != nil {
		return nil, fmt.Errorf("failed to execute request to CBR: %w", err)
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
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
			slog.Int("statusCode", resp.StatusCode),
			slog.String("body", string(body[:n])),
		)

		return nil, fmt.Errorf("CBR responded with unexpected HTTP status code: %d", resp.StatusCode)
	}

	// Dangerous conversion from Windows-1251 to UTF-8
	decoder := xml.NewDecoder(resp.Body)
	decoder.CharsetReader = indenticalCharsetReader

	var result RatesResponse
	err = decoder.Decode(&result)
	if err != nil {
		return nil, fmt.Errorf("failed to decode CBR response body: %w", err)
	}

	return &result, nil
}
