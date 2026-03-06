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

// F creates a Field with a single key-value pair.
// This is a convenience function to reduce allocations compared to map literals.
// Example: log.Info("message", gplog.F("user_id", 123), gplog.F("action", "login"))
func F(key string, value any) Field {
	return Field{Key: key, Value: value}
}

// M creates multiple fields from a map literal.
// Example: log.Info("message", gplog.M(map[string]any{"user_id": 123, "action": "login"})...)
func M(m map[string]any) []Field {
	fields := make([]Field, 0, len(m))
	for k, v := range m {
		fields = append(fields, Field{Key: k, Value: v})
	}
	return fields
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
