package main

import (
	"github.com/AlekseyPorandaykin/crypto_analyst/cmd"
	"github.com/AlekseyPorandaykin/crypto_analyst/pkg/logger"
	"go.uber.org/zap"
)

var version string

func main() {
	logger.InitDefaultLogger()
	defer func() { _ = zap.L().Sync() }()
	zap.L().Debug("Start app", zap.String("version", version))
	cmd.Execute()
}
