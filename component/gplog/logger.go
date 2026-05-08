package gplog

import "context"

const (
	JSONFormat    = "json"    // JSONFormat specifies the log format as JSON.
	ConsoleFormat = "console" // ConsoleFormat specifies the log format for console output.
)

// Logger defines the logging interface for the application.
// Usage example:
// 	log.Info("user logged in", "user_id", 123, "ip", "1.2.3.4")
type Logger interface {
	GetFormat() string
	Info(text string, keyvals ...any)
	Debug(text string, keyvals ...any)
	Warn(text string, keyvals ...any)
	Error(text string, keyvals ...any)
	Fatal(text string, keyvals ...any)

	InfoWithCtx(ctx context.Context, text string, keyvals ...any)
	DebugWithCtx(ctx context.Context, text string, keyvals ...any)
	WarnWithCtx(ctx context.Context, text string, keyvals ...any)
	ErrorWithCtx(ctx context.Context, text string, keyvals ...any)
	FatalWithCtx(ctx context.Context, text string, keyvals ...any)
}
