package handlers

import (
	"encoding/json"
	"encoding/xml"
	"fmt"
	"io"
	"log/slog"
	"net/http"
)

type Valute struct {
	ID        string `xml:"ID,attr" json:"id"`
	NumCode   string `json:"numCode"`
	CharCode  string `json:"charCode"`
	Nominal   uint   `json:"nominal"`
	Name      string `json:"name"`
	Value     string `json:"value"`
	VunitRate string `json:"vunitRate"`
}

type CbrCursResponse struct {
	Date       string   `xml:"Date,attr" json:"date"`
	Name       string   `xml:"name,attr" json:"name"`
	Currencies []Valute `xml:"Valute" json:"currencies"`
}

type ErrorResponse struct {
	Error string `json:"error"`
}

const getCursCbrMirrorUrl = "https://httpco.de/404" // "https://www.cbr-xml-daily.ru/daily_eng_utf8.xml"

func indenticalCharsetReader(encoding string, input io.Reader) (io.Reader, error) {
	return input, nil
}

func writeErr(w http.ResponseWriter, err error, statusCode int) {
	w.WriteHeader(statusCode)
	json.NewEncoder(w).Encode(ErrorResponse{err.Error()})
}

func HandleGetCurs(w http.ResponseWriter, r *http.Request, logger *slog.Logger) {
	const unexpectedStatusCodeResponseBodySizeLimit = 4 * 1024

	w.Header().Set("Content-Type", "application/json")

	resp, err := http.Get(getCursCbrMirrorUrl)
	if err != nil {
		writeErr(w, err, http.StatusInternalServerError)
		return
	}
	defer resp.Body.Close()

	if resp.StatusCode != http.StatusOK {
		body := make([]byte, unexpectedStatusCodeResponseBodySizeLimit)

		reader := io.LimitReader(resp.Body, unexpectedStatusCodeResponseBodySizeLimit)
		n, err := reader.Read(body)
		if n == 0 || err != nil {
			logger.Warn("Failed to get response body during processing unexpected HTTP status code",
				slog.Int("n", n),
				slog.Any("err", err),
			)
		}
		logger.Error("Received unexpected HTTP status code from CBR", slog.Int("statusCode", resp.StatusCode), slog.String("body", string(body[:n])))

		writeErr(
			w,
			fmt.Errorf("CBR responded with unexpected HTTP status code: %d", resp.StatusCode),
			http.StatusInternalServerError,
		)
		return
	}

	// Dangerous
	decoder := xml.NewDecoder(resp.Body)
	decoder.CharsetReader = indenticalCharsetReader

	parsedBody := CbrCursResponse{}
	err = decoder.Decode(&parsedBody)
	if err != nil {
		writeErr(w, err, http.StatusInternalServerError)
		return
	}

	encoder := json.NewEncoder(w)
	err = encoder.Encode(parsedBody)
	if err != nil {
		writeErr(w, err, http.StatusInternalServerError)
		return
	}
}
