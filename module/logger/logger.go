package logger

import "context"

// Logger defines a logging interface with methods for various log levels.
// Each method accepts a message and optional arguments for structured logging.
type Logger interface {
	// Info logs general informational messages.
	Info(msg string)

	// Warn logs warnings indicating potential issues that may not affect normal operation.
	Warn(msg string)

	// Debug logs detailed debugging information for troubleshooting purposes.
	Debug(msg string)

	// Error logs error messages for serious issues that require attention.
	Error(msg string)

	// ErrorWithCtx logs error messages with additional context, such as request-scoped data.
	ErrorWithCtx(ctx context.Context, msg string)
}
