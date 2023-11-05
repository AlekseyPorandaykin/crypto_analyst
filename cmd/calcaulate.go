package cmd

import (
	"context"
	"github.com/AlekseyPorandaykin/crypto_analyst/internal/price"
	"github.com/AlekseyPorandaykin/crypto_analyst/internal/repositories"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"os/signal"
	"syscall"
	"time"
)

var calculateCmd = &cobra.Command{
	Use: "calculate",
	Run: func(cmd *cobra.Command, args []string) {
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
			zap.L().Error("Error init database", zap.Error(err))
			return
		}
		defer func() { _ = db.Close() }()

		symbolRepo := repositories.NewSymbols(db)
		priceChangesRepo := repositories.NewPriceChanges(db)
		calc := price.NewCalculate(symbolRepo, priceChangesRepo)
		if err := calc.ReportAvg(ctx, time.Now().Add(-10*24*time.Hour), time.Now()); err != nil {
			zap.L().Error("Error calculate", zap.Error(err))
		}
	},
}

func init() {
	rootCmd.AddCommand(calculateCmd)
}
