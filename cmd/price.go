package cmd

import (
	"context"
	"errors"
	"fmt"
	price2 "github.com/AlekseyPorandaykin/crypto_analyst/internal/components/price"
	"github.com/AlekseyPorandaykin/crypto_analyst/internal/repositories"
	http_server "github.com/AlekseyPorandaykin/crypto_analyst/internal/server/http"
	"github.com/AlekseyPorandaykin/crypto_analyst/internal/storage"
	"github.com/AlekseyPorandaykin/crypto_analyst/internal/storage/cache"
	"github.com/AlekseyPorandaykin/crypto_loader/api/http/client"
	"github.com/spf13/cobra"
	"net/http"
	"os/signal"
	"syscall"
	"time"
)

var priceCmd = &cobra.Command{
	Use: "price",
	Run: func(cmd *cobra.Command, args []string) {
		const DefaultRecalculateDuration = 5 * time.Minute
		const DefaultPriceAggregationDuration = 1 * time.Hour
		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()
		db, err := repositories.CreateDB(repositories.Config{
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
		defer func() { _ = db.Close() }()
		priceRepo := repositories.NewPriceRepository(db)
		priceChangesRepo := repositories.NewPriceChanges(db)
		symbolRepo := repositories.NewSymbols(db)
		aggregationRepo := repositories.NewAggregation(db)

		calculatorApp := price2.NewChangeCalculator(priceRepo, priceChangesRepo, symbolRepo)

		priceStorage := storage.NewComposite(cache.NewPrice(), priceRepo)

		loaderApp, err := client.NewClient("http://localhost:8081", http.DefaultClient)
		if err != nil {
			fmt.Println("Error init loader: ", err.Error())
			return
		}
		loaderPrice := price2.NewLoader(loaderApp, priceStorage)
		metricCalculator := price2.NewMetricCalculator(priceChangesRepo, aggregationRepo, symbolRepo)

		priceController := price2.NewController(priceStorage)
		serv := http_server.NewServer()
		serv.Registration(priceController)

		go func() {
			defer cancel()
			if err := loaderPrice.Run(ctx); err != nil && errors.Is(err, context.Canceled) {
				fmt.Printf("error execute loader price: %s \n", err.Error())
			}
		}()
		go func() {
			defer cancel()
			if err := calculatorApp.Run(ctx, DefaultRecalculateDuration); err != nil {
				fmt.Printf("error execute app: %s \n", err.Error())
			}
		}()
		go func() {
			defer cancel()
			if err := serv.Run(":8082"); err != nil {
				fmt.Println("error execute server: ", err.Error())
			}
		}()

		go metricCalculator.Run(ctx, DefaultPriceAggregationDuration)

		<-ctx.Done()
	},
}

func init() {
	rootCmd.AddCommand(priceCmd)
}
