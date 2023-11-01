package cmd

import (
	"context"
	"errors"
	"github.com/spf13/cobra"
	"go.uber.org/zap"
)

var rootCmd = &cobra.Command{Use: "crypto-analyst"}
var PriceCmd = &cobra.Command{
	Use:   "price",
	Short: "Price analyst",
}

func init() {
	rootCmd.AddCommand(PriceCmd)
}

func Execute() {
	if err := rootCmd.Execute(); err != nil && !errors.Is(err, context.Canceled) {
		zap.L().Error("execute root cmd", zap.Error(err))
	}
}
