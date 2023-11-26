package main

import (
	"github.com/AlekseyPorandaykin/crypto_analyst/cmd"
	"github.com/AlekseyPorandaykin/crypto_analyst/pkg/logger"
	"github.com/prometheus/client_golang/prometheus/promhttp"
	"go.uber.org/zap"
	"net/http"
)

var version string

func main() {
	logger.InitDefaultLogger()
	defer func() { _ = zap.L().Sync() }()
	zap.L().Debug("Start app", zap.String("version", version))
	go func() {
		http.Handle("/metrics", promhttp.Handler())
		if err := http.ListenAndServe(":9082", nil); err != nil {
			zap.L().Fatal("error start metric", zap.Error(err))
		}
	}()
	cmd.Execute()
}
