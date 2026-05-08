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
func Info(text string, keyvals ...any) {
	if defaultLogger == nil {
		panic("application log is not initialized")
	}
	defaultLogger.Info(text, keyvals...)
}

// Debug logs a debug level message.
func Debug(text string, keyvals ...any) {
	if defaultLogger == nil {
		panic("application log is not initialized")
	}
	defaultLogger.Debug(text, keyvals...)
}

// Warn logs a warning level message.
func Warn(text string, keyvals ...any) {
	if defaultLogger == nil {
		panic("application log is not initialized")
	}
	defaultLogger.Warn(text, keyvals...)
}

// Error logs an error level message.
func Error(text string, keyvals ...any) {
	if defaultLogger == nil {
		panic("application log is not initialized")
	}
	defaultLogger.Error(text, keyvals...)
}

// Fatal logs a fatal level message and exits the application.
func Fatal(text string, keyvals ...any) {
	if defaultLogger == nil {
		panic("application log is not initialized")
	}
	defaultLogger.Fatal(text, keyvals...)
}

// InfoWithCtx logs an info level message with keyvals.
func InfoWithCtx(ctx context.Context, text string, keyvals ...any) {
	if defaultLogger == nil {
		panic("application log is not initialized")
	}
	defaultLogger.InfoWithCtx(ctx, text, keyvals...)
}

// DebugWithCtx logs a debug level message with keyvals.
func DebugWithCtx(ctx context.Context, text string, keyvals ...any) {
	if defaultLogger == nil {
		panic("application log is not initialized")
	}
	defaultLogger.DebugWithCtx(ctx, text, keyvals...)
}

// WarnWithCtx logs a warning level message with keyvals.
func WarnWithCtx(ctx context.Context, text string, keyvals ...any) {
	if defaultLogger == nil {
		panic("application log is not initialized")
	}
	defaultLogger.WarnWithCtx(ctx, text, keyvals...)
}

// ErrorWithCtx logs an error level message with keyvals.
func ErrorWithCtx(ctx context.Context, text string, keyvals ...any) {
	if defaultLogger == nil {
		panic("application log is not initialized")
	}
	defaultLogger.ErrorWithCtx(ctx, text, keyvals...)
}

// FatalWithCtx logs a fatal level message with keyvals and exits the application.
func FatalWithCtx(ctx context.Context, text string, keyvals ...any) {
	if defaultLogger == nil {
		panic("application log is not initialized")
	}
	defaultLogger.FatalWithCtx(ctx, text, keyvals...)
}