package cmd

import (
	"context"
	"fmt"
	"github.com/AlekseyPorandaykin/crypto_analyst/internal/price"
	"github.com/AlekseyPorandaykin/crypto_analyst/internal/repositories"
	"github.com/AlekseyPorandaykin/crypto_analyst/internal/server/http"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
	"os/signal"
	"syscall"
)

var httpServerCmd = &cobra.Command{
	Use:   "http",
	Short: "Run http web server",
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
		serv := http.NewServer(calc)
		go func() {
			defer cancel()
			if err := serv.Run(":8082"); err != nil {
				fmt.Println("error execute server: ", err.Error())
			}
		}()
		fmt.Println("server run")
		<-ctx.Done()
	},
}

func init() {
	ServerCmd.AddCommand(httpServerCmd)
}
