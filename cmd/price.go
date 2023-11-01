package cmd

import (
	"context"
	"errors"
	"fmt"
	"github.com/AlekseyPorandaykin/crypto_analyst/internal/client/loader"
	"github.com/AlekseyPorandaykin/crypto_analyst/internal/price"
	"github.com/AlekseyPorandaykin/crypto_analyst/internal/repositories"
	"github.com/spf13/cobra"
	"os/signal"
	"syscall"
	"time"
)

var priceCmd = &cobra.Command{
	Use: "price",
	Run: func(cmd *cobra.Command, args []string) {
		const DefaultRecalculateDuration = time.Hour
		ctx, cancel := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
		defer cancel()
		db, err := repositories.CreateDB(repositories.Config{
			Driver:   "postgres",
			Username: "crypto_analyst",
			Password: "developer",
			Host:     "localhost",
			Port:     "5434",
			Database: "crypto_analyst",
		})
		if err != nil {
			fmt.Println("Error init database: ", err.Error())
			return
		}
		defer func() { _ = db.Close() }()
		priceRepo := repositories.NewPriceRepository(db)
		priceChangesRepo := repositories.NewPriceChanges(db)
		ap := price.NewPrice(priceRepo, priceChangesRepo)

		l := loader.NewLoader("localhost:50052")
		loaderPrice := price.NewLoader(l, priceRepo)
		go func() {
			defer cancel()
			if err := loaderPrice.Run(ctx); err != nil && errors.Is(err, context.Canceled) {
				fmt.Printf("error execute loader price: %s \n", err.Error())
			}
		}()
		go func() {
			defer cancel()
			if err := l.Start(ctx, 60); err != nil && errors.Is(err, context.Canceled) {
				fmt.Printf("error execute loader: %s \n", err.Error())
			}
		}()
		go func() {
			defer cancel()
			if err := ap.Run(ctx, DefaultRecalculateDuration); err != nil {
				fmt.Printf("error execute app: %s \n", err.Error())
			}
		}()

		<-ctx.Done()
	},
}

func init() {
	rootCmd.AddCommand(priceCmd)
}
