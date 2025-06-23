package gplog

import (
	"context"

	"github.com/archine/gin-plus/v4/component/gplog/iface"
	"github.com/archine/gin-plus/v4/internal/logger"
)

func Info(text string) {
	logger.GlobalLogger.Info(text)
}

func Debug(text string) {
	logger.GlobalLogger.Debug(text)
}

func Warn(text string) {
	logger.GlobalLogger.Warn(text)
}

func Error(text string) {
	logger.GlobalLogger.Error(text)
}

func Fatal(text string) {
	logger.GlobalLogger.Fatal(text)
}

func WithContext(ctx context.Context) iface.AbstractAppLogger {
	return logger.GlobalLogger.WithContext(ctx)
}
