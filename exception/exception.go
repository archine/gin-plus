package exception

import (
	"fmt"
	"strings"
)

// Wrap formats an error message by wrapping the provided error with additional gpctx.
// It supports including multiple additional errors in the format string.
//
// Parameters:
// - err: The primary error to be wrapped.
// - msg: A message to add gpctx to the primary error. If empty, no additional message is included.
// - more: Additional errors to be appended to the format string for more gpctx.
//
// Returns: a formatted error that includes the primary error, optional message, and any additional errors.
//
// Example usage:
//
//	err := errors.New("original error")
//	wrappedErr := Wrap(err, "additional gpctx", anotherErr)
//	fmt.Println(wrappedErr) // Output: original error: additional gpctx: anotherErr
func Wrap(err error, msg string, more ...error) error {
	if err == nil {
		return nil
	}
	var builder strings.Builder
	args := make([]any, 0, 2+len(more))

	// Append the first error and message (if any)
	if msg != "" {
		builder.WriteString("%s")
		args = append(args, msg)
	}
	builder.WriteString(": %w")
	args = append(args, err)

	// Append any additional errors
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

// WrapF formats an error message by wrapping the provided error with additional gpctx.
// It supports including multiple additional errors in the format string.
//
// Parameters:
// - err: The primary error to be wrapped.
// - format: A format specifier for the message to add gpctx to the primary error.
// - args: Arguments to be formatted into the message.
//
// Returns: a formatted error that includes the primary error, formatted message, and any additional errors.
//
// Example usage:
//
//	err := errors.New("original error")
//	wrappedErr := WrapF(err, "additional gpctx: %v", anotherErr)
//	fmt.Println(wrappedErr) // Output: original error: additional gpctx: anotherErr
func WrapF(err error, format string, args ...any) error {
	return Wrap(err, fmt.Sprintf(format, args...))
}
