package internal

import (
	"fmt"
	"github.com/archine/gin-plus/v3/module/logger"
	"os"
)

// Logger is the global logger instance, initially set to a default logger.
var Logger logger.Logger = &DefaultLogger{}

// DefaultLogger provides a basic implementation of Logger.
type DefaultLogger struct{}

func (d *DefaultLogger) Info(msg string, args ...any) {
	fmt.Printf("[INFO] "+msg+"\n", args...)
}

func (d *DefaultLogger) Warn(msg string, args ...any) {
	fmt.Printf("[WARN] "+msg+"\n", args...)
}

func (d *DefaultLogger) Debug(msg string, args ...any) {
	fmt.Printf("[DEBUG] "+msg+"\n", args...)
}

func (d *DefaultLogger) Error(msg string, args ...any) {
	fmt.Printf("[ERROR] "+msg+"\n", args...)
}

func (d *DefaultLogger) Fatal(msg string, args ...any) {
	fmt.Printf("[FATAL] "+msg+"\n", args...)
	os.Exit(1)
}
