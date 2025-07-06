package gplog

import (
	"context"
)

var (
	// globalLog is the global logger instance used throughout the application.
	globalLog Logger
)


// set by the application during initialization.
func setLogger(l Logger) {
	globalLog = l
}

// Field represents a key-value pair for structured logging.
type Field struct {
	Key   string
	Value any
}

// Logger defines the logging interface for the application.
type Logger interface {
	Info(text string, fields ...Field)
	Debug(text string, fields ...Field)
	Warn(text string, fields ...Field)
	Error(text string, fields ...Field)
	Fatal(text string, fields ...Field)

	InfoWithCtx(ctx context.Context, text string, fields ...Field)
	DebugWithCtx(ctx context.Context, text string, fields ...Field)
	WarnWithCtx(ctx context.Context, text string, fields ...Field)
	ErrorWithCtx(ctx context.Context, text string, fields ...Field)
	FatalWithCtx(ctx context.Context, text string, fields ...Field)
}

// Info logs an info level message.
func Info(text string, fields ...Field) {
	if globalLog == nil {
		panic("globalLog is not initialized")
	}
	globalLog.Info(text, fields...)
}

// Debug logs a debug level message.
func Debug(text string, fields ...Field) {
	if globalLog == nil {
		panic("globalLog is not initialized")
	}
	globalLog.Debug(text, fields...)
}

// Warn logs a warning level message.
func Warn(text string, fields ...Field) {
	if globalLog == nil {
		panic("globalLog is not initialized")
	}
	globalLog.Warn(text, fields...)
}

// Error logs an error level message.
func Error(text string, fields ...Field) {
	if globalLog == nil {
		panic("globalLog is not initialized")
	}
	globalLog.Error(text, fields...)
}

// Fatal logs a fatal level message and exits the application.
func Fatal(text string, fields ...Field) {
	if globalLog == nil {
		panic("globalLog is not initialized")
	}
	globalLog.Fatal(text, fields...)
}

// InfoWithCtx logs an info level message with context fields.
func InfoWithCtx(ctx context.Context, text string, fields ...Field) {
	if globalLog == nil {
		panic("globalLog is not initialized")
	}
	globalLog.InfoWithCtx(ctx, text, fields...)
}

// DebugWithCtx logs a debug level message with context fields.
func DebugWithCtx(ctx context.Context, text string, fields ...Field) {
	if globalLog == nil {
		panic("globalLog is not initialized")
	}
	globalLog.DebugWithCtx(ctx, text, fields...)
}

// WarnWithCtx logs a warning level message with context fields.
func WarnWithCtx(ctx context.Context, text string, fields ...Field) {
	if globalLog == nil {
		panic("globalLog is not initialized")
	}
	globalLog.WarnWithCtx(ctx, text, fields...)
}

// ErrorWithCtx logs an error level message with context fields.
func ErrorWithCtx(ctx context.Context, text string, fields ...Field) {
	if globalLog == nil {
		panic("globalLog is not initialized")
	}
	globalLog.ErrorWithCtx(ctx, text, fields...)
}

// FatalWithCtx logs a fatal level message with context fields and exits the application.
func FatalWithCtx(ctx context.Context, text string, fields ...Field) {
	if globalLog == nil {
		panic("globalLog is not initialized")
	}
	globalLog.FatalWithCtx(ctx, text, fields...)
}
