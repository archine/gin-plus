package gplog

import "context"

const (
	JSONFormat    = "json"    // JSONFormat specifies the log format as JSON.
	ConsoleFormat = "console" // ConsoleFormat specifies the log format for console output.
)

// Field represents a key-value pair for structured logging.
type Field struct {
	Key   string
	Value any
}

// F is a shorthand function to create a Field.
func F(key string, value any) Field {
	return Field{Key: key, Value: value}
}

// Logger defines the logging interface for the application.
type Logger interface {
	GetFormat() string
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
