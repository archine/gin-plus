package logger

var (
	Log AbstractLogger // Log logger instance
)

type AbstractLogger interface {
	// Init logger
	// the application will call this method automatically.
	// you can also call this method manually.
	Init()

	// GetLogger returns the corresponding log collector
	GetLogger() any

	// Infof logs a message at InfoLevel, args are handled in the manner of fmt.sprintf.
	Infof(msg string, args ...any)

	// Warnf logs a message at WarnLevel, args are handled in the manner of fmt.sprintf.
	Warnf(msg string, args ...any)

	// Debugf logs a message at DebugLevel, args are handled in the manner of fmt.sprintf.
	Debugf(msg string, args ...any)

	// Errorf logs a message at ErrorLevel, args are handled in the manner of fmt.sprintf.
	Errorf(msg string, args ...any)

	// Info logs a message at InfoLevel
	Info(v ...any)

	// Warn logs a message at WarnLevel
	Warn(v ...any)

	// Debug logs a message at DebugLevel
	Debug(v ...any)

	// Error logs a message at ErrorLevel
	Error(v ...any)

	// Println logs a message at InfoLevel
	Println(v ...any)

	// Printf logs a message at InfoLevel
	Printf(format string, v ...any)

	// Fatal logs a message at FatalLevel
	Fatal(v ...any)

	// Fatalf Fatal logs a message at FatalLevel
	Fatalf(format string, v ...any)
}
