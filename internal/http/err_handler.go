package http

import (
	"errors"
	"log/slog"
	"net/http"

	"github.com/labstack/echo/v5"
)

type ErrorResponse struct {
	Error string `json:"error"`
}

func httpErrorHandler(c *echo.Context, err error) {
	if resp, uErr := echo.UnwrapResponse(c.Response()); uErr == nil {
		if resp.Committed {
			return // already sent by a handler/middleware
		}
	}

	code := http.StatusInternalServerError
	var sc echo.HTTPStatusCoder
	if errors.As(err, &sc) {
		if tmp := sc.StatusCode(); tmp != 0 {
			code = tmp
		}
	}

	cErr := c.JSON(code, ErrorResponse{Error: err.Error()})
	if cErr != nil {
		c.Logger().Error("Failed to send error response", slog.Any("error", errors.Join(err, cErr)))
	}
}
