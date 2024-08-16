package logger

// Logger defines the interface for logging at different levels.
type Logger interface {
	Info(msg string, args ...any)
	Warn(msg string, args ...any)
	Debug(msg string, args ...any)
	Error(msg string, args ...any)
	Fatal(msg string, args ...any)
}
