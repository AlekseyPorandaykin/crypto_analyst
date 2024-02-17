package controller

import (
	"fmt"
	"github.com/AlekseyPorandaykin/crypto_analyst/domain"
	"github.com/AlekseyPorandaykin/crypto_analyst/dto"
	"github.com/AlekseyPorandaykin/crypto_analyst/internal/components/controller/templates"
	"github.com/AlekseyPorandaykin/crypto_analyst/internal/storage/db"
	"github.com/labstack/echo/v4"
	"github.com/pkg/errors"
	"go.uber.org/zap"
	"html/template"
	"io"
	"net/http"
	"sort"
	"time"
)

type Price struct {
	priceStorage      *db.PriceRepository
	snapshotStorage   domain.CandlestickStorage
	symbolStorage     domain.SymbolStorage
	priceChangeLoader domain.PriceChangeLoader
}

func NewPrice(
	priceStorage *db.PriceRepository,
	snapshotStorage domain.CandlestickStorage,
	symbolStorage domain.SymbolStorage,
	priceChangeLoader domain.PriceChangeLoader,
) *Price {
	return &Price{
		priceStorage:      priceStorage,
		snapshotStorage:   snapshotStorage,
		symbolStorage:     symbolStorage,
		priceChangeLoader: priceChangeLoader,
	}
}

func (app *Price) RegistrationRoute(e *echo.Echo) {
	e.GET("/price", app.index)
	e.GET("/price/:symbol", app.symbolPrice)
	e.GET("/price/:exchange/:symbol/changes", app.changes)
	e.GET("/price/snapshot/:exchange/:symbol", app.snapshot)
	e.GET("/api/price/new", app.newPrices)
}

func (app *Price) index(c echo.Context) error {
	symbols, err := app.symbolStorage.List(c.Request().Context())
	if err != nil {
		return err
	}
	sort.Slice(symbols, func(i, j int) bool {
		return domain.PopularSymbols[symbols[i]] > domain.PopularSymbols[symbols[j]]
	})

	return executeTemplate("index", templates.IndexHtmlPage, c.Response(), templates.PageData{Title: "Price", Data: symbols})
}

func (app *Price) symbolPrice(c echo.Context) error {
	symbol := c.Param("symbol")
	if symbol == "" {
		return errors.New("empty symbol")
	}
	prices, err := app.priceStorage.Prices(c.Request().Context(), symbol)
	if err != nil {
		return err
	}
	return executeTemplate("symbol_price", templates.SymbolHtmlPage, c.Response(), templates.PageData{Title: "Symbol " + symbol, Symbol: symbol, Data: prices})

}

func (app *Price) snapshot(c echo.Context) error {
	symbol := c.Param("symbol")
	if symbol == "" {
		return errors.New("empty symbol")
	}
	exchange := c.Param("exchange")
	if exchange == "" {
		return errors.New("empty exchange")
	}
	snapshotOneHour, err := app.snapshotStorage.LastCandlestick(c.Request().Context(), exchange, symbol, domain.OneHourInterval)
	if err != nil {
		return err
	}
	snapshotFourHour, err := app.snapshotStorage.LastCandlestick(c.Request().Context(), exchange, symbol, domain.FourHourInterval)
	snapshots := []*dto.Candlestick{snapshotFourHour, snapshotOneHour}
	if err != nil {
		return err
	}
	data := map[string]interface{}{
		"time":      time.Now().In(time.UTC),
		"snapshots": snapshots,
	}
	return c.JSON(http.StatusOK, data)
}

func (app *Price) changes(c echo.Context) error {
	symbol := c.Param("symbol")
	if symbol == "" {
		return errors.New("empty symbol")
	}
	exchange := c.Param("exchange")
	if exchange == "" {
		return errors.New("empty exchange")
	}
	data, err := app.priceChangeLoader.Changes(
		c.Request().Context(), exchange, symbol, time.Now().Add(-24*time.Hour), time.Now(),
	)
	if err != nil {
		return err
	}
	return executeTemplate("price_change", templates.PriceChangesHtmlPage, c.Response(), templates.PageData{Title: fmt.Sprintf("%s-%s", exchange, symbol), Symbol: symbol, Data: data})
}

func (app *Price) newPrices(c echo.Context) error {
	symbols, err := app.priceStorage.NewSymbols(c.Request().Context(), time.Now().Add(-24*time.Hour))
	if err != nil {
		zap.L().Error("controller new prices", zap.Error(err))
		return c.JSON(http.StatusInternalServerError, nil)
	}
	return c.JSON(http.StatusOK, symbols)
}

func createTemplate(name string, content []byte) (*template.Template, error) {
	var text []byte
	text = append(text, templates.HeaderHtml...)
	text = append(text, content...)
	text = append(text, templates.FooterHtml...)
	return template.New(name).Parse(string(text))
}

func executeTemplate(name string, content []byte, wr io.Writer, data templates.PageData) error {
	templ, err := createTemplate(name, content)
	if err != nil {
		return err
	}
	data.CurrentTime = time.Now().In(time.UTC)
	return templ.ExecuteTemplate(wr, name, data)
}
