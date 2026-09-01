package cbr

import (
	"context"
	"io"
	"log/slog"
	"net/http"
	"reflect"
	"strings"
	"testing"
)

type myRoundTripper struct {
	roundTripFunc func(*http.Request) (*http.Response, error)
}

func (rt *myRoundTripper) RoundTrip(r *http.Request) (*http.Response, error) {
	return rt.roundTripFunc(r)
}

type myUnexpectedEofReader struct{}

func (r *myUnexpectedEofReader) Read(p []byte) (n int, err error) {
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
	Currencies: []Currency{
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
			&myRoundTripper{roundTripFunc: func(r *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode:    http.StatusOK,
					Header:        http.Header{"Content-Type": []string{"text/xml"}},
					Body:          io.NopCloser(strings.NewReader(ratesXml)),
					ContentLength: int64(len(ratesXml)),
					Request:       r,
				}, nil
			}},
			ratesExpectedResult,
			false,
		},
		{
			"invalid xml response",
			args{
				defCtx,
			},
			&myRoundTripper{roundTripFunc: func(r *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode:    http.StatusOK,
					Header:        http.Header{"Content-Type": []string{"text/xml"}},
					Body:          io.NopCloser(strings.NewReader(invalidRatesXml)),
					ContentLength: int64(len(invalidRatesXml)),
					Request:       r,
				}, nil
			}},
			nil,
			true,
		},
		{
			"valid xml response but invalid schema",
			args{
				defCtx,
			},
			&myRoundTripper{roundTripFunc: func(r *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode:    http.StatusOK,
					Header:        http.Header{"Content-Type": []string{"text/xml"}},
					Body:          io.NopCloser(strings.NewReader(validButNotRatesXml)),
					ContentLength: int64(len(validButNotRatesXml)),
					Request:       r,
				}, nil
			}},
			nil,
			true,
		},
		{
			"json response",
			args{
				defCtx,
			},
			&myRoundTripper{roundTripFunc: func(r *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode:    http.StatusOK,
					Header:        http.Header{"Content-Type": []string{"application/json"}},
					Body:          io.NopCloser(strings.NewReader(ratesJson)),
					ContentLength: int64(len(ratesJson)),
					Request:       r,
				}, nil
			}},
			nil,
			true,
		},
		{
			"html response",
			args{
				defCtx,
			},
			&myRoundTripper{roundTripFunc: func(r *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode:    http.StatusOK,
					Header:        http.Header{"Content-Type": []string{"text/html"}},
					Body:          io.NopCloser(strings.NewReader(sampleHtml)),
					ContentLength: int64(len(sampleHtml)),
					Request:       r,
				}, nil
			}},
			nil,
			true,
		},
		{
			"empty response",
			args{
				defCtx,
			},
			&myRoundTripper{roundTripFunc: func(r *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode:    http.StatusOK,
					Header:        http.Header{"Content-Type": []string{"text/xml"}},
					Body:          io.NopCloser(strings.NewReader("")),
					ContentLength: 0,
					Request:       r,
				}, nil
			}},
			nil,
			true,
		},
		{
			"interrupted xml response",
			args{
				defCtx,
			},
			&myRoundTripper{roundTripFunc: func(r *http.Request) (*http.Response, error) {
				return &http.Response{
					StatusCode: http.StatusOK,
					Header:     http.Header{"Content-Type": []string{"text/xml"}},
					Body: io.NopCloser(io.MultiReader(
						io.LimitReader(strings.NewReader(ratesXml), int64(len(ratesXml)/2)),
						&myUnexpectedEofReader{}),
					),
					ContentLength: int64(len(ratesXml)),
					Request:       r,
				}, nil
			}},
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

//func TestClient_GetRatesTransport(t *testing.T) {
//	defCtx := context.Background()
//	logger := slog.New(slog.DiscardHandler)
//
//	type args struct {
//		ctx context.Context
//	}
//	tests := []struct {
//		name      string
//		args      args
//		transport http.RoundTripper
//		want      *RatesResponse
//		wantErr   bool
//	}{
//		{
//			"valid xml response",
//			args{
//				context.TODO(),
//			},
//			...,
//
//		},
//		// TODO: Add test cases.
//	}
//
//	httptest.NewServer()
//
//	for _, tt := range tests {
//		t.Run(tt.name, func(t *testing.T) {
//			c := &Client{
//				httpClient: &http.Client{Transport: tt.transport},
//				logger:     logger,
//			}
//			got, err := c.GetRates(tt.args.ctx)
//			if (err != nil) != tt.wantErr {
//				t.Errorf("GetRates() error = %v, wantErr %v", err, tt.wantErr)
//				return
//			}
//			if !reflect.DeepEqual(got, tt.want) {
//				t.Errorf("GetRates() got = %v, want %v", got, tt.want)
//			}
//		})
//	}
//}

//func TestClient_GetRatesTimeout(t *testing.T) {
//
//}
