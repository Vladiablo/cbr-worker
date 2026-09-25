package currency

import (
	"cbr-worker/internal/cbr/service"
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v5"
)

type Handler struct {
	svc    *service.Service
	logger *slog.Logger
}

func New(svc *service.Service, logger *slog.Logger) *Handler {
	return &Handler{svc: svc, logger: logger}
}

// GetRates
//
//	@Summary		Get rates
//	@Description	Get list of CBR exchange rates
//	@Tags			rates
//	@Produce		json
//	@Param			currency	query		[]string	false	"List of currencies to get exchange rates for. By default, get exchange rates for all available currencies"		CollectionFormat(csv)	Example(AUD,USD)
//	@Param			fromDate	query		string		false	"ISO 8601 formatted date specifying the start date of exchange rates list. Must be used together with `toDate`"	Format(date)			Example(2006-01-02)
//	@Param			toDate		query		string		false	"ISO 8601 formatted date specifying the end date of exchange rates list. Must be used together with `fromDate`"	Format(date)			Example(2006-01-31)
//	@Success		200			{array}		cbr.ExchangeRates
//	@Failure		400			{object}	http.ErrorResponse
//	@Failure		404			{object}	http.ErrorResponse
//	@Failure		500			{object}	http.ErrorResponse
//	@Router			/rates [get]
func (h *Handler) GetRates(c *echo.Context) error {
	var qp QueryParams
	if err := qp.Parse(c); err != nil {
		return echo.ErrBadRequest.Wrap(err)
	}

	params := &service.GetRatesParams{
		Currencies: qp.Currencies,
		FromDate:   qp.FromDate.Time,
		ToDate:     qp.ToDate.Time,
	}

	rates, err := h.svc.GetRates(c.Request().Context(), params)
	if err != nil {
		if err, ok := errors.AsType[service.GetRatesParamsError](err); ok {
			return echo.ErrBadRequest.Wrap(err)
		} else {
			h.logger.Error("Failed to get exchange rates", slog.Any("error", err))

			return echo.ErrInternalServerError.Wrap(err)
		}
	}

	if len(rates) == 0 {
		return echo.NewHTTPError(http.StatusNotFound, "no exchange rates found")
	}

	return c.JSON(http.StatusOK, rates)
}
