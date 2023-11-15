package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/AlekseyPorandaykin/crypto_analyst/internal/price"
	"github.com/AlekseyPorandaykin/crypto_analyst/internal/repositories"
	"github.com/AlekseyPorandaykin/crypto_loader/api/grpc/client"
	"github.com/spf13/cobra"
	"os/signal"
	"syscall"
	"time"
)

var priceCmd = &cobra.Command{
	Use: "price",
	Run: func(cmd *cobra.Command, args []string) {
		const DefaultRecalculateDuration = 5 * time.Minute
		const DefaultPriceAggregationDuration = 1 * time.Hour
		const DefaultLoadPriceDurationSec = 60
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

		calculatorApp := price.NewChangeCalculator(priceRepo, priceChangesRepo, symbolRepo)

		loaderApp := client.NewLoader("localhost:50052")
		loaderPrice := price.NewLoader(loaderApp, priceRepo)
		metricCalculator := price.NewMetricCalculator(priceChangesRepo, aggregationRepo, symbolRepo)

		go func() {
			defer cancel()
			if err := loaderPrice.Run(ctx); err != nil && errors.Is(err, context.Canceled) {
				fmt.Printf("error execute loader price: %s \n", err.Error())
			}
		}()
		go func() {
			defer cancel()
			if err := loaderApp.Start(ctx, DefaultLoadPriceDurationSec); err != nil && errors.Is(err, context.Canceled) {
				fmt.Printf("error execute loader: %s \n", err.Error())
			}
		}()
		go func() {
			defer cancel()
			if err := calculatorApp.Run(ctx, DefaultRecalculateDuration); err != nil {
				fmt.Printf("error execute app: %s \n", err.Error())
			}
		}()

		go metricCalculator.Run(ctx, DefaultPriceAggregationDuration)

		<-ctx.Done()
	},
}

func init() {
	rootCmd.AddCommand(priceCmd)
}
