package price

import (
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"net/http"
	"time"
)

const DatetimeFormat = "2006-01-02 15:04:05"

type Controller struct {
	calculate *Calculate
}

func NewController(calculate *Calculate) *Controller {
	return &Controller{calculate: calculate}
}

func (app *Controller) RegistrationRoute(e *echo.Echo) {
	e.GET("/price", app.priceCoefficient)
}

func (app *Controller) priceCoefficient(c echo.Context) error {
	params := &struct {
		Symbol string `json:"symbol" query:"symbol"`
		From   string `json:"from" query:"from"`
		To     string `json:"to" query:"to"`
	}{}
	if err := c.Bind(params); err != nil {
		return err
	}
	if params.Symbol == "" {
		return errors.New("empty symbol")
	}
	from, err := time.Parse(DatetimeFormat, params.From)
	if err != nil {
		return err
	}
	to, err := time.Parse(DatetimeFormat, params.To)
	if err != nil {
		return err
	}
	res, err := app.calculate.ReportPriceChanges(
		c.Request().Context(),
		params.Symbol,
		from.In(time.UTC),
		to.In(time.UTC))
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, res)
}
