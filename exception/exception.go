package exception

import (
	"fmt"
	"strings"
)

// Wrap wraps the original error, optionally attaches a custom message, and supports chaining multiple additional errors.
// Args:
//   - err: The original error to wrap.
//   - msg: Optional custom message to attach.
//   - more: Optional additional errors to chain.
//
// Returns: A formatted error containing the original error, custom message, and all additional errors.
// Example:
//
//	err := errors.New("original error")
//	wrappedErr := Wrap(err, "extra info", anotherErr)
//	fmt.Println(wrappedErr) // Output: original error: extra info: anotherErr
func Wrap(err error, msg string, more ...error) error {
	if err == nil {
		return nil
	}
	var builder strings.Builder
	args := make([]any, 0, 2+len(more))

	if msg != "" {
		builder.WriteString("%s")
		args = append(args, msg)
	}
	builder.WriteString(": %w")
	args = append(args, err)

	for _, e := range more {
		builder.WriteString(": %w")
		args = append(args, e)
	}

	if builder.Len() == 0 {
		// If no valid message was added, return the original error
		return err
	}
	return fmt.Errorf(builder.String(), args...)
}

// WrapF wraps the original error with a formatted message and supports chaining multiple additional errors.
// Args:
//   - err: The original error to wrap.
//   - format: Format string for the custom message.
//   - args: Arguments for formatting the message.
//
// Returns: A formatted error containing the original error and the formatted message.
// Example:
//
//	err := errors.New("original error")
//	wrappedErr := WrapF(err, "extra info: %v", anotherErr)
//	fmt.Println(wrappedErr) // Output: original error: extra info: anotherErr
func WrapF(err error, format string, args ...any) error {
	return Wrap(err, fmt.Sprintf(format, args...))
}
