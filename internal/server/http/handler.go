package http

import (
	"github.com/AlekseyPorandaykin/crypto_analyst/internal/price"
	"github.com/labstack/echo/v4"
	"net/http"
	"time"
)

const DatetimeFormat = "2006-01-02 15:04:05"

type Handler struct {
	calculate *price.Calculate
}

func NewHandler(calculate *price.Calculate) *Handler {
	return &Handler{calculate: calculate}
}

func (h *Handler) RegistrationRoute(e *echo.Echo) {
	e.GET("/price", h.priceCoefficient)
}

func (h *Handler) priceCoefficient(c echo.Context) error {
	params := &struct {
		Symbol string `json:"symbol" query:"symbol"`
		From   string `json:"from" query:"from"`
		To     string `json:"to" query:"to"`
	}{}
	if err := c.Bind(params); err != nil {
		return err
	}
	from, err := time.Parse(DatetimeFormat, params.From)
	if err != nil {
		return err
	}
	to, err := time.Parse(DatetimeFormat, params.To)
	if err != nil {
		return err
	}
	res, err := h.calculate.ReportPriceChanges(
		c.Request().Context(),
		params.Symbol,
		from.In(time.UTC),
		to.In(time.UTC))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
}
