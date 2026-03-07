package exception

import (
	"fmt"
	"strings"
)

// Wrap wraps an error with an optional context message and supports chaining multiple errors.
// It returns nil if the input error is nil, making it safe to use in error propagation chains.
//
// Parameters:
//   - err: The error to wrap (returns nil if err is nil)
//   - msg: Optional context message (can be empty string)
//   - more: Optional additional errors to chain together
//
// Returns: A wrapped error that preserves the error chain for errors.Is and errors.As.
//
// Examples:
//
//	// Basic wrapping with context
//	err := errors.New("connection failed")
//	wrapped := Wrap(err, "database operation failed")
//	// Output: database operation failed: connection failed
//
//	// Wrapping without message
//	wrapped := Wrap(err, "")
//	// Output: connection failed
//
//	// Chaining multiple errors
//	err1 := errors.New("timeout")
//	err2 := errors.New("retry failed")
//	wrapped := Wrap(err1, "operation failed", err2)
//	// Output: operation failed: timeout: retry failed
//
//	// Safe nil handling
//	wrapped := Wrap(nil, "some context")
//	// Output: nil
func Wrap(err error, msg string, more ...error) error {
	if err == nil {
		return nil
	}
	var builder strings.Builder
	args := make([]any, 0, 2+len(more))

	if msg != "" {
		builder.WriteString("%s: ")
		args = append(args, msg)
	}
	builder.WriteString("%w")
	args = append(args, err)

	for _, e := range more {
		builder.WriteString(": %w")
		args = append(args, e)
	}

	return fmt.Errorf(builder.String(), args...)
}

// WrapF wraps an error with a formatted context message using fmt.Sprintf formatting.
// It's a convenience wrapper around Wrap that allows printf-style formatting.
//
// Parameters:
//   - err: The error to wrap (returns nil if err is nil)
//   - format: Printf-style format string for the context message
//   - args: Arguments for the format string
//
// Returns: A wrapped error with the formatted message.
//
// Examples:
//
//	// Format with variables
//	err := errors.New("not found")
//	wrapped := WrapF(err, "user %s (id: %d) lookup failed", "zhangsan", 123)
//	// Output: user zhangsan (id: 123) lookup failed: not found
//
//	// Format with error values
//	err1 := errors.New("timeout")
//	err2 := errors.New("connection refused")
//	wrapped := WrapF(err1, "failed after %d retries: %v", 3, err2)
//	// Output: failed after 3 retries: connection refused: timeout
func WrapF(err error, format string, args ...any) error {
	return Wrap(err, fmt.Sprintf(format, args...))
}
