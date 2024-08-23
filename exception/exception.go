package exception

import (
	"fmt"
	"github.com/archine/gin-plus/v3/module/constant/errs"
	"github.com/archine/gin-plus/v3/module/stacktrace"
	"strings"
)

// BusinessException represents a service-level exception that includes stack trace information.
// The default business error code is set to 40000, which corresponds to bcode.BadRequest.
//
// When returned via resp.DirectRespErr, this error is not treated as an unknown error,
// so resp.ServerError will not be triggered.
type BusinessException struct {
	code int    // Error code representing the specific business error.
	msg  string // Error message describing the exception.
}

// Error returns the error message.
func (b *BusinessException) Error() string {
	return b.msg
}

// Code returns the business error code.
func (b *BusinessException) Code() int {
	return b.code
}

// NewBusinessErr creates a new BusinessException with a default error code of 40000.
func NewBusinessErr(msg string) *BusinessException {
	return &BusinessException{code: errs.BadRequestErr.Code(), msg: msg}
}

// NewBusinessErrWithCode creates a new BusinessException with a specified error code.
func NewBusinessErrWithCode(code int, msg string) *BusinessException {
	return &BusinessException{code: code, msg: msg}
}

// StackError represents an exception that includes stack trace information.
//
// When formatting the error with fmt.Printf:
// - Using format '%+v' will include the stack trace information in the output.
// - Using other formats will display only the error message.
//
// Usage example:
//
//	err := NewStackErr("error message")
//	fmt.Printf("%+v", err)
type StackError struct {
	msg string // Error message describing the exception.
	st  string // Stack trace information.
}

func (s *StackError) Error() string {
	return s.msg
}

func (s *StackError) Format(f fmt.State, verb rune) {
	switch verb {
	case 'v':
		if f.Flag('+') {
			_, _ = fmt.Fprintf(f, "%s\n%s", s.msg, s.st)
			return
		}
		fallthrough
	default:
		_, _ = fmt.Fprintf(f, "%s", s.msg)
	}
}

// StackTrace returns the stack trace of the error.
func (s *StackError) StackTrace() string {
	return s.st
}

// NewStackErr creates a new StackError with the specified message.
func NewStackErr(msg string) *StackError {
	stack := stacktrace.Capture(0, 1)
	defer stack.Free()
	return &StackError{msg, stack.First()}
}

// WithStack wraps an error with a stack trace.
func WithStack(err error) *StackError {
	if err == nil {
		return nil
	}
	stack := stacktrace.Capture(0, 1)
	defer stack.Free()
	return &StackError{err.Error(), stack.First()}
}

// Wrap formats an error message by wrapping the provided error with additional context.
// It supports including multiple additional errors in the format string.
//
// Parameters:
// - err: The primary error to be wrapped.
// - msg: A message to add context to the primary error. If empty, no additional message is included.
// - more: Additional errors to be appended to the format string for more context.
//
// Returns: a formatted error that includes the primary error, optional message, and any additional errors.
//
// Example usage:
//
//	err := errors.New("original error")
//	wrappedErr := Wrap(err, "additional context", anotherErr)
//	fmt.Println(wrappedErr) // Output: original error: additional context: anotherErr
func Wrap(err error, msg string, more ...error) error {
	if err == nil {
		return nil
	}
	var builder strings.Builder
	args := make([]any, 0, 2+len(more))

	// Append the first error and message (if any)
	builder.WriteString("%w")
	args = append(args, err)
	if msg != "" {
		builder.WriteString(": %s")
		args = append(args, msg)
	}

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

// WrapF formats an error message by wrapping the provided error with additional context.
// It supports including multiple additional errors in the format string.
//
// Parameters:
// - err: The primary error to be wrapped.
// - format: A format specifier for the message to add context to the primary error.
// - args: Arguments to be formatted into the message.
//
// Returns: a formatted error that includes the primary error, formatted message, and any additional errors.
//
// Example usage:
//
//	err := errors.New("original error")
//	wrappedErr := WrapF(err, "additional context: %v", anotherErr)
//	fmt.Println(wrappedErr) // Output: original error: additional context: anotherErr
func WrapF(err error, format string, args ...any) error {
	return Wrap(err, fmt.Sprintf(format, args...))
}
