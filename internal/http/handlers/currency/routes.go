package currency

import "github.com/labstack/echo/v5"

func RegisterRoutes(e *echo.Group, currencyHandler *Handler) {
	grp := e.Group("/rates")
	grp.GET("", currencyHandler.GetRates)
}
