package shutdown

import (
	"os"
	"runtime/debug"
	"time"

	"go.uber.org/zap"
)

func HandlePanic() {
	if err := recover(); err != nil {
		zap.L().Error(
			"handle panic",
			zap.Any("recover", err),
			zap.ByteString("stack", debug.Stack()),
		)
		time.Sleep(5 * time.Second)
		os.Exit(1)
	}
}
