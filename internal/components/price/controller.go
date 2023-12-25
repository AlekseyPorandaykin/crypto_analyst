package price

import (
	"github.com/AlekseyPorandaykin/crypto_analyst/internal/storage"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"net/http"
)

const DatetimeFormat = "2006-01-02 15:04:05"

type Controller struct {
	loader storage.PriceLoader
}

func NewController(loader storage.PriceLoader) *Controller {
	return &Controller{loader: loader}
}

func (app *Controller) RegistrationRoute(e *echo.Echo) {
	e.GET("/price/:symbol", app.priceCoefficient)
}

func (app *Controller) priceCoefficient(c echo.Context) error {
	symbol := c.Param("symbol")
	if symbol == "" {
		return errors.New("empty symbol")
	}
	prices, err := app.loader.Prices(c.Request().Context(), symbol)
	if err != nil {
		return err
	}
	return c.JSON(http.StatusOK, prices)
}
