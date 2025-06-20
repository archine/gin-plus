package iface

import "context"

// AbstractAppLogger is an interface that defines the methods for logging in the application.
type AbstractAppLogger interface {
	Info(text string)
	Debug(text string)
	Warn(text string)
	Error(text string)
	Fatal(text string)
	// WithContext returns a new logger with the context.
	// The context can be used to add additional fields to the log output.
	WithContext(ctx context.Context) AbstractAppLogger
}
