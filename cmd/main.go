package main

import (
	"cbr-worker/internal/cbr"
	"log/slog"
	"net/http"
	"os"
)

// TODO: Place Client inside Handler Struct
// Client encapsulates loggger and HTTP client
// Handler struct encapsulates client and calls its method
// Use custom HTTP server & client

func main() {
	logger := slog.New(slog.NewJSONHandler(os.Stderr, nil))
	cbr := cbr.CbrClient{Logger: logger}

	http.HandleFunc("GET /api/v1/getCurs", cbr.HandleGetCurs)

	logger.Info("Starting HTTP server at port 8080...")
	err := http.ListenAndServe(":8080", nil)
	if err != nil {
		logger.Error("Failed to start HTTP server. Exiting...", slog.Any("error", err))
		os.Exit(1)
	}
}
