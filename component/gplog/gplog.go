package gplog

import (
	"context"
)

// defaultLogger is installed once by App.Run before any bean or server starts,
// so later reads from other goroutines need no synchronization.
var defaultLogger Logger

// SetLogger sets the global logger.
func SetLogger(logger Logger) {
	defaultLogger = logger
}

// GetLogger returns the global logger.
// It panics with an explicit hint if called before the logger is initialized.
func GetLogger() Logger {
	if defaultLogger == nil {
		panic("gplog: logger is not initialized yet, it becomes available after the application " +
			"loads its configuration in App.Run; do not log in init() or config/logger factories")
	}
	return defaultLogger
}

// Info logs an info level message.
func Info(text string, keyvals ...any) { GetLogger().Info(text, keyvals...) }

// Debug logs a debug level message.
func Debug(text string, keyvals ...any) { GetLogger().Debug(text, keyvals...) }

// Warn logs a warning level message.
func Warn(text string, keyvals ...any) { GetLogger().Warn(text, keyvals...) }

// Error logs an error level message.
func Error(text string, keyvals ...any) { GetLogger().Error(text, keyvals...) }

// Fatal logs a fatal level message and exits the application.
func Fatal(text string, keyvals ...any) { GetLogger().Fatal(text, keyvals...) }

// InfoWithCtx logs an info level message with keyvals.
func InfoWithCtx(ctx context.Context, text string, keyvals ...any) {
	GetLogger().InfoWithCtx(ctx, text, keyvals...)
}

// DebugWithCtx logs a debug level message with keyvals.
func DebugWithCtx(ctx context.Context, text string, keyvals ...any) {
	GetLogger().DebugWithCtx(ctx, text, keyvals...)
}

// WarnWithCtx logs a warning level message with keyvals.
func WarnWithCtx(ctx context.Context, text string, keyvals ...any) {
	GetLogger().WarnWithCtx(ctx, text, keyvals...)
}

// ErrorWithCtx logs an error level message with keyvals.
func ErrorWithCtx(ctx context.Context, text string, keyvals ...any) {
	GetLogger().ErrorWithCtx(ctx, text, keyvals...)
}

// FatalWithCtx logs a fatal level message with keyvals and exits the application.
func FatalWithCtx(ctx context.Context, text string, keyvals ...any) {
	GetLogger().FatalWithCtx(ctx, text, keyvals...)
}
