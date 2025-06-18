package gplog

import (
	"context"

	"github.com/archine/gin-plus/v3/internal/logger"
	"github.com/archine/gin-plus/v3/module/gplog/iface"
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
