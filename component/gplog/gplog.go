package gplog

import "context"

// defaultLogger holds the global logger instance (private).
var defaultLogger Logger

// SetLogger sets the global logger instance.
// This should only be called by the framework during initialization.
func SetLogger(logger Logger) {
	defaultLogger = logger
}

// GetLogger returns the global log instance.
func GetLogger() Logger {
	if defaultLogger == nil {
		panic("application log is not initialized")
	}
	return defaultLogger
}

// Info logs an info level message.
func Info(text string, fields ...Field) {
	if defaultLogger == nil {
		panic("application log is not initialized")
	}
	defaultLogger.Info(text, fields...)
}

// Debug logs a debug level message.
func Debug(text string, fields ...Field) {
	if defaultLogger == nil {
		panic("application log is not initialized")
	}
	defaultLogger.Debug(text, fields...)
}

// Warn logs a warning level message.
func Warn(text string, fields ...Field) {
	if defaultLogger == nil {
		panic("application log is not initialized")
	}
	defaultLogger.Warn(text, fields...)
}

// Error logs an error level message.
func Error(text string, fields ...Field) {
	if defaultLogger == nil {
		panic("application log is not initialized")
	}
	defaultLogger.Error(text, fields...)
}

// Fatal logs a fatal level message and exits the application.
func Fatal(text string, fields ...Field) {
	if defaultLogger == nil {
		panic("application log is not initialized")
	}
	defaultLogger.Fatal(text, fields...)
}

// InfoWithCtx logs an info level message with fields.
func InfoWithCtx(ctx context.Context, text string, fields ...Field) {
	if defaultLogger == nil {
		panic("application log is not initialized")
	}
	defaultLogger.InfoWithCtx(ctx, text, fields...)
}

// DebugWithCtx logs a debug level message with fields.
func DebugWithCtx(ctx context.Context, text string, fields ...Field) {
	if defaultLogger == nil {
		panic("application log is not initialized")
	}
	defaultLogger.DebugWithCtx(ctx, text, fields...)
}

// WarnWithCtx logs a warning level message with fields.
func WarnWithCtx(ctx context.Context, text string, fields ...Field) {
	if defaultLogger == nil {
		panic("application log is not initialized")
	}
	defaultLogger.WarnWithCtx(ctx, text, fields...)
}

// ErrorWithCtx logs an error level message with fields.
func ErrorWithCtx(ctx context.Context, text string, fields ...Field) {
	if defaultLogger == nil {
		panic("application log is not initialized")
	}
	defaultLogger.ErrorWithCtx(ctx, text, fields...)
}

// FatalWithCtx logs a fatal level message with fields and exits the application.
func FatalWithCtx(ctx context.Context, text string, fields ...Field) {
	if defaultLogger == nil {
		panic("application log is not initialized")
	}
	defaultLogger.FatalWithCtx(ctx, text, fields...)
}