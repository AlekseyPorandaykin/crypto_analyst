package shutdown

import (
	"go.uber.org/zap"
	"os"
	"time"
)

func HandlePanic() {
	if err := recover(); err != nil {
		zap.L().Error("handle panic", zap.Any("recover", err))
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}
}
