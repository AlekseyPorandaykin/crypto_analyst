package cmd

import (
	"context"
	"errors"
	"fmt"
	"net"
	"net/http"
	"os/signal"
	"syscall"
	"time"

	"github.com/AlekseyPorandaykin/crypto_analyst/internal/components/calculation"
	"github.com/AlekseyPorandaykin/crypto_analyst/internal/components/controller"
	"github.com/AlekseyPorandaykin/crypto_analyst/internal/components/loader"
	"github.com/AlekseyPorandaykin/crypto_analyst/internal/storage"
	"github.com/AlekseyPorandaykin/crypto_analyst/internal/storage/cache"
	"github.com/AlekseyPorandaykin/crypto_analyst/internal/storage/db"
	"github.com/AlekseyPorandaykin/crypto_analyst/pkg/database"
	http_server "github.com/AlekseyPorandaykin/crypto_analyst/pkg/server/http"
	"github.com/AlekseyPorandaykin/crypto_analyst/pkg/shutdown"
	"github.com/AlekseyPorandaykin/crypto_loader/api/http/client"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var rootCmd = &cobra.Command{
	Use: "price",
	Run: func(cmd *cobra.Command, args []string) {
		const DefaultRecalculateDuration = 5 * time.Second
		const DefaultPriceAggregationDuration = 1 * time.Hour
		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()
		connect, err := database.CreateConnection(database.Config{
			Driver:   "postgres",
			Username: "crypto_app",
			Password: "developer",
			Host:     "localhost",
			Port:     "5433",
			Database: "crypto_app",
		})
		if err != nil {
			fmt.Println("Error init database: ", err.Error())
			return
		}
		defer func() { _ = connect.Close() }()
		priceRepo := db.NewPriceRepository(connect)
		priceChangesRepo := db.NewPriceChanges(connect)

		symbolRepo := db.NewSymbols(connect)
		aggregationRepo := db.NewAggregation(connect)

		calculatorApp := calculation.NewChangeCalculator(priceRepo, priceChangesRepo, symbolRepo)

		priceStorage := storage.NewComposite(cache.NewPrice(), priceRepo)

		loaderApp, err := client.NewClient("http://localhost:8081", http.DefaultClient)
		if err != nil {
			fmt.Println("Error init loader: ", err.Error())
			return
		}
		candlestickRepo := db.NewCandlestick(connect)
		candlestickCache := cache.NewCandlestick()
		candlestickStorage := storage.NewCandlestickComposite(candlestickCache, candlestickRepo)

		price := loader.NewPrice(loaderApp, symbolRepo, priceRepo)
		loaderPrice := loader.NewLoader(loaderApp, priceStorage, candlestickStorage, price)
		metricCalculator := calculation.NewChangeCoefficient(priceChangesRepo, aggregationRepo, symbolRepo)

		techAnalysis := calculation.NewTechAnalysis(candlestickStorage)

		priceController := controller.NewPrice(priceRepo, candlestickStorage, symbolRepo, priceChangesRepo)
		if err != nil {
			fmt.Println("Error init price controller: ", err.Error())
			return
		}
		serv := http_server.NewServer()
		defer serv.Close()
		serv.RegistrationPage(priceController)
		serv.RegistrationApi(priceController)
		serv.WithAuthor("developer")
		serv.WithApplicationName("crypto_analyst")

		go func() {
			defer shutdown.HandlePanic()
			defer cancel()
			if err := loaderPrice.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
				fmt.Printf("error execute loader price: %s \n", err.Error())
			}
		}()
		go func() {
			defer shutdown.HandlePanic()
			defer cancel()
			if err := calculatorApp.Run(ctx, DefaultRecalculateDuration); err != nil && !errors.Is(err, context.Canceled) {
				fmt.Printf("error execute app: %s \n", err.Error())
			}
		}()
		go func() {
			defer shutdown.HandlePanic()
			defer cancel()
			if err := serv.Run(net.JoinHostPort("localhost", "8082")); err != nil {
				fmt.Println("error execute server: ", err.Error())
			}
		}()
		go func() {
			defer shutdown.HandlePanic()
			defer cancel()
			if err := techAnalysis.Run(ctx); err != nil && !errors.Is(err, context.Canceled) {
				fmt.Println("error execute techAnalysis: ", err.Error())
			}
		}()

		go func() {
			defer shutdown.HandlePanic()
			metricCalculator.Run(ctx, DefaultPriceAggregationDuration)
		}()

		<-ctx.Done()
	},
}

func Execute() {
	if err := rootCmd.Execute(); err != nil && !errors.Is(err, context.Canceled) {
		zap.L().Error("execute root cmd", zap.Error(err))
	}
}
