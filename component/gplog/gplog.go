package gplog

import (
	"context"
)

// Field represents a key-value pair for logging.
type Field struct {
	Key   string
	Value any
}

// Logger is an interface that defines the methods for logging in the app.
type Logger interface {
	Info(text string, fields ...Field)

	Debug(text string, fields ...Field)

	Warn(text string, fields ...Field)

	Error(text string, fields ...Field)

	Fatal(text string, fields ...Field)

	// WithContext returns a new syslog with the context.
	// The context can be used to add additional fields to the syslog output.
	WithContext(ctx context.Context) Logger
}

var defaultLogger Logger

// SetDefaultLogger sets the default logger implementation
func SetDefaultLogger(logger Logger) {
	if logger == nil {
		return
	}
	defaultLogger = logger
}

// Info logs an info level message
func Info(text string, fields ...Field) {
	defaultLogger.Info(text, fields...)
}

// Debug logs a debug level message
func Debug(text string, fields ...Field) {
	defaultLogger.Debug(text, fields...)
}

// Warn logs a warning level message
func Warn(text string, fields ...Field) {
	defaultLogger.Warn(text, fields...)
}

// Error logs an error level message
func Error(text string, fields ...Field) {
	defaultLogger.Error(text, fields...)
}

// Fatal logs a fatal level message and exits
func Fatal(text string, fields ...Field) {
	defaultLogger.Fatal(text, fields...)
}

// WithContext returns a syslog with context values
func WithContext(ctx context.Context) Logger {
	return defaultLogger.WithContext(ctx)
}
