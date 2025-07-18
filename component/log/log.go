package log

import (
	"context"
	"github.com/archine/gin-plus/v4/component/log/logcore"
	"github.com/archine/gin-plus/v4/internal/vars/syslog"
)

// Field is an alias for Field. Aliasing this type dramatically
// improves the navigability of this package's API documentation.
type Field = logcore.Field

// GetLogger returns the global log instance.
func GetLogger() logcore.Logger {
	if syslog.Log == nil {
		panic("application log is not initialized")
	}
	return syslog.Log
}

// Info logs an info level message.
func Info(text string, fields ...Field) {
	if syslog.Log == nil {
		panic("application log is not initialized")
	}
	syslog.Log.Info(text, fields...)
}

// Debug logs a debug level message.
func Debug(text string, fields ...Field) {
	if syslog.Log == nil {
		panic("application log is not initialized")
	}
	syslog.Log.Debug(text, fields...)
}

// Warn logs a warning level message.
func Warn(text string, fields ...Field) {
	if syslog.Log == nil {
		panic("application log is not initialized")
	}
	syslog.Log.Warn(text, fields...)
}

// Error logs an error level message.
func Error(text string, fields ...Field) {
	if syslog.Log == nil {
		panic("application log is not initialized")
	}
	syslog.Log.Error(text, fields...)
}

// Fatal logs a fatal level message and exits the application.
func Fatal(text string, fields ...Field) {
	if syslog.Log == nil {
		panic("application log is not initialized")
	}
	syslog.Log.Fatal(text, fields...)
}

// InfoWithCtx logs an info level message with fields.
func InfoWithCtx(ctx context.Context, text string, fields ...Field) {
	if syslog.Log == nil {
		panic("application log is not initialized")
	}
	syslog.Log.InfoWithCtx(ctx, text, fields...)
}

// DebugWithCtx logs a debug level message with  fields.
func DebugWithCtx(ctx context.Context, text string, fields ...Field) {
	if syslog.Log == nil {
		panic("application log is not initialized")
	}
	syslog.Log.DebugWithCtx(ctx, text, fields...)
}

// WarnWithCtx logs a warning level message with fields.
func WarnWithCtx(ctx context.Context, text string, fields ...Field) {
	if syslog.Log == nil {
		panic("application log is not initialized")
	}
	syslog.Log.WarnWithCtx(ctx, text, fields...)
}

// ErrorWithCtx logs an error level message with fields.
func ErrorWithCtx(ctx context.Context, text string, fields ...Field) {
	if syslog.Log == nil {
		panic("application log is not initialized")
	}
	syslog.Log.ErrorWithCtx(ctx, text, fields...)
}

// FatalWithCtx logs a fatal level message with fields and exits the application.
func FatalWithCtx(ctx context.Context, text string, fields ...Field) {
	if syslog.Log == nil {
		panic("application log is not initialized")
	}
	syslog.Log.FatalWithCtx(ctx, text, fields...)
}
