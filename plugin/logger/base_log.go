package logger

var (
	Log AbstractLogger = &DefaultLog{} // Log logger instance
)

type AbstractLogger interface {
	// Info logs a message at InfoLevel, args are handled in the manner of fmt.sprintf.
	Info(msg string, args ...any)

	// Warn logs a message at WarnLevel, args are handled in the manner of fmt.sprintf.
	Warn(msg string, args ...any)

	// Debug logs a message at DebugLevel, args are handled in the manner of fmt.sprintf.
	Debug(msg string, args ...any)

	// Error logs a message at ErrorLevel, args are handled in the manner of fmt.sprintf.
	Error(msg string, args ...any)

	// Fatal logs a message at FatalLevel
	Fatal(format string, v ...any)
}
