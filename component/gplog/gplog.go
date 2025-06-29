package gplog

import (
	"context"
	"sync"
)

var (
	globalLog Logger
	once      sync.Once
)

// setLogger the global logger for the application.
func setLogger(l Logger) {
	once.Do(func() {
		globalLog = l
	})
}

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

// Info logs an info level message
func Info(text string, fields ...Field) {
	globalLog.Info(text, fields...)
}

// Debug logs a debug level message
func Debug(text string, fields ...Field) {
	globalLog.Debug(text, fields...)
}

// Warn logs a warning level message
func Warn(text string, fields ...Field) {
	globalLog.Warn(text, fields...)
}

// Error logs an error level message
func Error(text string, fields ...Field) {
	globalLog.Error(text, fields...)
}

// Fatal logs a fatal level message and exits
func Fatal(text string, fields ...Field) {
	globalLog.Fatal(text, fields...)
}

// WithContext returns a syslog with context values
func WithContext(ctx context.Context) Logger {
	return globalLog.WithContext(ctx)
}
