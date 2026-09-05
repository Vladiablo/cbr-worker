package cbr

import (
	"context"
	"fmt"
	"io"
	"log/slog"
	"net"
	"net/http"
	"reflect"
	"strings"
	"testing"
	"time"
)

type myRoundTripper struct {
	roundTripFunc func(*http.Request) (*http.Response, error)
}

func (rt *myRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return rt.roundTripFunc(r)
}

type myUnexpectedEofReader struct{}

func (r *myUnexpectedEofReader) Read(_ []byte) (n int, err error) {
	return 0, io.ErrUnexpectedEOF
}

const ratesXml = `<?xml version="1.0" encoding="windows-1251"?>
<ValCurs Date="01.09.2026" name="Foreign Currency Market">
<Valute ID="R01010">
<NumCode>036</NumCode>
<CharCode>AUD</CharCode>
<Nominal>1</Nominal>
<Name>Australian Dollar</Name>
<Value>61,8994</Value>
<VunitRate>61,8994</VunitRate>
</Valute>
<Valute ID="R01020A">
<NumCode>944</NumCode>
<CharCode>AZN</CharCode>
<Nominal>1</Nominal>
<Name>Azerbaijan Manat</Name>
<Value>50,8114</Value>
<VunitRate>50,8114</VunitRate>
</Valute>
<Valute ID="R01030">
<NumCode>012</NumCode>
<CharCode>DZD</CharCode>
<Nominal>100</Nominal>
<Name>Algerian Dinar</Name>
<Value>64,9383</Value>
<VunitRate>0,649383</VunitRate>
</Valute>
</ValCurs>`

const invalidRatesXml = `<?xml version="1.0" encoding="windows-1251"?>
<ValCurs Date="01.09.2026" name="Foreign Currency Market">
<Valute ID="R01010">
<NumCode>036</NumCode>
<CharCode>AUD</CharCode>
<Nominal>1
<Name>Australian Dollar</Name>
<Value>61,8994</Value>
<VunitRate>61,8994</VunitRate>
</Valute>
<Valute ID="R01020A">
<NumCode>944</
<VunitRate>50,8114</VunitRate>
</Valute>
<Valute ID="R01030">
<VunitRate>0,649383</VunitRate>
</Valute>`

const validButNotRatesXml = `<?xml version="1.0" encoding="windows-1251"?>
<bookstore>
  <book category="fiction">
    <title>The Great Adventure</title>
    <author>Jane Doe</author>
    <year>2023</year>
    <price>19.99</price>
  </book>
  <book category="science">
    <title>The Living Earth</title>
    <author>John Smith</author>
    <year>2021</year>
    <price>29.50</price>
  </book>
</bookstore>`

const ratesJson = `{
  "date": "01.09.2026",
  "name": "Foreign Currency Market",
  "currencies": [
    {
      "id": "R01010",
      "numCode": "036",
      "charCode": "AUD",
      "nominal": 1,
      "name": "Australian Dollar",
      "value": "61,8994",
      "vunitRate": "61,8994"
    },
    {
      "id": "R01020A",
      "numCode": "944",
      "charCode": "AZN",
      "nominal": 1,
      "name": "Azerbaijan Manat",
      "value": "50,8114",
      "vunitRate": "50,8114"
    },
    {
      "id": "R01030",
      "numCode": "012",
      "charCode": "DZD",
      "nominal": 100,
      "name": "Algerian Dinar",
      "value": "64,9383",
      "vunitRate": "0,649383"
    }
  ]
}`

const sampleHtml = `<!DOCTYPE html>
<html lang="en">
<head>
    <title>Sample HTML Page</title>
</head>
<body>
    <header>
        <h1>Welcome to My Website</h1>
        <p>A simple, clean template to kickstart your web design.</p>
    </header>
</body>
</html>`

var ratesExpectedResult = &RatesResponse{
	Date: "01.09.2026",
	Name: "Foreign Currency Market",
	Currencies: []*RawCurrency{
		{
			ID:        "R01010",
			NumCode:   "036",
			CharCode:  "AUD",
			Nominal:   1,
			Name:      "Australian Dollar",
			Value:     "61,8994",
			VunitRate: "61,8994",
		},
		{
			ID:        "R01020A",
			NumCode:   "944",
			CharCode:  "AZN",
			Nominal:   1,
			Name:      "Azerbaijan Manat",
			Value:     "50,8114",
			VunitRate: "50,8114",
		},
		{
			ID:        "R01030",
			NumCode:   "012",
			CharCode:  "DZD",
			Nominal:   100,
			Name:      "Algerian Dinar",
			Value:     "64,9383",
			VunitRate: "0,649383",
		},
	},
}

type RoundTripperParams struct {
	statusCode    int
	contentType   string
	stringBody    string
	body          io.Reader
	contentLength int
}

func createRoundTripper(cfg *RoundTripperParams) http.RoundTripper {
	sc := cfg.statusCode
	if sc == 0 {
		sc = http.StatusOK
	}

	header := make(http.Header)
	if len(cfg.contentType) != 0 {
		header["Content-Type"] = []string{cfg.contentType}
	}

	cl := cfg.contentLength
	body := cfg.body
	if body == nil {
		body = strings.NewReader(cfg.stringBody)
		if cl == 0 {
			cl = len(cfg.stringBody)
		}
	}

	return &myRoundTripper{roundTripFunc: func(r *http.Request) (*http.Response, error) {
		return &http.Response{
			StatusCode:    sc,
			Header:        header,
			Body:          io.NopCloser(body),
			ContentLength: int64(cl),
			Request:       r,
		}, nil
	}}
}

func TestClient_GetRates(t *testing.T) {
	defCtx := context.Background()
	logger := slog.New(slog.DiscardHandler)

	type args struct {
		ctx context.Context
	}
	tests := []struct {
		name      string
		args      args
		transport http.RoundTripper
		want      *RatesResponse
		wantErr   bool
	}{
		{
			"valid xml response",
			args{
				defCtx,
			},
			createRoundTripper(&RoundTripperParams{contentType: "text/xml", stringBody: ratesXml}),
			ratesExpectedResult,
			false,
		},
		{
			"invalid xml response",
			args{
				defCtx,
			},
			createRoundTripper(&RoundTripperParams{contentType: "text/xml", stringBody: invalidRatesXml}),
			nil,
			true,
		},
		{
			"valid xml response but invalid schema",
			args{
				defCtx,
			},
			createRoundTripper(&RoundTripperParams{contentType: "text/xml", stringBody: validButNotRatesXml}),
			nil,
			true,
		},
		{
			"json response",
			args{
				defCtx,
			},
			createRoundTripper(&RoundTripperParams{contentType: "application/json", stringBody: ratesJson}),
			nil,
			true,
		},
		{
			"html response",
			args{
				defCtx,
			},
			createRoundTripper(&RoundTripperParams{contentType: "text/html", stringBody: sampleHtml}),
			nil,
			true,
		},
		{
			"empty response",
			args{
				defCtx,
			},
			createRoundTripper(&RoundTripperParams{contentType: "text/xml", stringBody: ""}),
			nil,
			true,
		},
		{
			"interrupted xml response",
			args{
				defCtx,
			},
			createRoundTripper(&RoundTripperParams{
				contentType: "text/xml",
				body: io.MultiReader(
					io.LimitReader(strings.NewReader(ratesXml), int64(len(ratesXml)/2)),
					&myUnexpectedEofReader{},
				),
				contentLength: len(ratesXml),
			}),
			nil,
			true,
		},
		{
			"not found response",
			args{
				defCtx,
			},
			createRoundTripper(&RoundTripperParams{statusCode: http.StatusNotFound,
				contentType: "text/plain", stringBody: "Not Found"},
			),
			nil,
			true,
		},
		{
			"internal server error response",
			args{
				defCtx,
			},
			createRoundTripper(&RoundTripperParams{statusCode: http.StatusInternalServerError,
				contentType: "text/plain", stringBody: "Internal Server Error"},
			),
			nil,
			true,
		},
		{
			"failed connection",
			args{
				defCtx,
			},
			&http.Transport{
				DialContext: func(ctx context.Context, network string, addr string) (net.Conn, error) {
					return nil, fmt.Errorf("dial tcp: connection refused")
				},
			},
			nil,
			true,
		},
	}

	for _, tt := range tests {
		t.Run(tt.name, func(t *testing.T) {
			c := &Client{
				httpClient: &http.Client{Transport: tt.transport},
				logger:     logger,
			}
			got, err := c.GetRates(tt.args.ctx)
			if (err != nil) != tt.wantErr {
				t.Errorf("GetRates() error = %v, wantErr %v", err, tt.wantErr)
				return
			}
			if !reflect.DeepEqual(got, tt.want) {
				t.Errorf("GetRates() got = %v, want %v", got, tt.want)
			}
		})
	}
}

func TestClient_GetRatesContext(t *testing.T) {
	defCtx := context.Background()
	logger := slog.New(slog.DiscardHandler)

	roundTripperWithCtx := &myRoundTripper{
		roundTripFunc: func(r *http.Request) (*http.Response, error) {
			reqCtx := r.Context()
			<-reqCtx.Done()
			return nil, context.Cause(reqCtx)
		},
	}

	t.Run("context timeout", func(t *testing.T) {
		c := &Client{
			httpClient: &http.Client{Transport: roundTripperWithCtx},
			logger:     logger,
		}
		ctx, cancel := context.WithTimeout(defCtx, 0)
		defer cancel()

		_, err := c.GetRates(ctx)
		if err == nil {
			t.Errorf("GetRates() error = %v, wantErr %v", err, true)
		}
	})

	t.Run("context canceled", func(t *testing.T) {
		c := &Client{
			httpClient: &http.Client{Transport: roundTripperWithCtx},
			logger:     logger,
		}
		ctx, cancel := context.WithCancel(defCtx)
		cancel()

		_, err := c.GetRates(ctx)
		if err == nil {
			t.Errorf("GetRates() error = %v, wantErr %v", err, true)
		}
	})

	t.Run("http client timeout", func(t *testing.T) {
		c := &Client{
			httpClient: &http.Client{Transport: roundTripperWithCtx, Timeout: time.Microsecond},
			logger:     logger,
		}

		_, err := c.GetRates(defCtx)
		if err == nil {
			t.Errorf("GetRates() error = %v, wantErr %v", err, true)
		}
	})
}
