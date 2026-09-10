package currency

import (
	"cbr-worker/internal/cbr/service"
	"encoding/json"
	"errors"
	"fmt"
	"log/slog"
	"net/http"
)

type Handler struct {
	svc    *service.Service
	logger *slog.Logger
}
type ErrorResponse struct {
	Error string `json:"error"`
}

func New(svc *service.Service, logger *slog.Logger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

func (h *Handler) writeErr(w http.ResponseWriter, err error, statusCode int) {
	w.WriteHeader(statusCode)
	err = json.NewEncoder(w).Encode(ErrorResponse{err.Error()})
	if err != nil {
		h.logger.Error("Failed to encode error response", slog.Any("error", err))
	}
}

func (h *Handler) GetRates(w http.ResponseWriter, r *http.Request) {
	w.Header().Set("Content-Type", "application/json")

	qp := &QueryParams{}
	if err := qp.Parse(r.URL.Query()); err != nil {
		h.writeErr(w, err, http.StatusBadRequest)

		return
	}

	params := &service.GetRatesParams{
		Currencies: qp.currencies,
		FromDate:   qp.fromDate,
		ToDate:     qp.toDate,
	}

	rates, err := h.svc.GetRates(r.Context(), params)
	if err != nil {
		if err, ok := errors.AsType[service.GetRatesParamsError](err); ok {
			h.writeErr(w, err, http.StatusBadRequest)
		} else {
			h.logger.Error("Failed to get exchange rates", slog.Any("error", err))
			h.writeErr(w, err, http.StatusInternalServerError)
		}

		return
	}

	if len(rates) == 0 {
		h.writeErr(w, fmt.Errorf("no exchange rates found"), http.StatusNotFound)

		return
	}

	err = json.NewEncoder(w).Encode(rates)
	if err != nil {
		h.writeErr(w, err, http.StatusInternalServerError)
	}
}
