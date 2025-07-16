package gplogcore

import "context"

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
