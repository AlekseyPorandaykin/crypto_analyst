package main

import (
	"github.com/AlekseyPorandaykin/crypto_analyst/cmd"
	"github.com/AlekseyPorandaykin/crypto_analyst/pkg/logger"
	"github.com/AlekseyPorandaykin/crypto_analyst/pkg/metrics"
	"go.uber.org/zap"
)

var version string

func main() {
	logger.InitDefaultLogger()
	defer func() { _ = zap.L().Sync() }()
	zap.L().Debug("Start app", zap.String("version", version))
	go func() {
		if err := metrics.Handler("localhost", "9082"); err != nil {
			zap.L().Fatal("error start metric", zap.Error(err))
		}
	}()
	cmd.Execute()
}
