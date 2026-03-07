package exception

import (
	"fmt"

	"github.com/archine/gin-plus/v4/exception/stacktrace"
)

// CodedError is an interface for errors that carry a business error code.
type CodedError interface {
	error
	Code() int
}

// BusinessException represents a service-level exception that includes stack trace information.
// The default business error code is set to DefaultBusinessCode, which corresponds to bcode.BadRequest.
//
// When returned via resp.DirectRespErr, this error is not treated as an unknown error,
// so resp.ServerError will not be triggered.
type BusinessException struct {
	code int    // Business error code, default is DefaultBusinessCode.
	msg  string // Error message describing the exception.
}

func (b *BusinessException) Error() string {
	return b.msg
}

func (b *BusinessException) Code() int {
	return b.code
}

// WithMessage returns a new BusinessException with the updated message.
func (b *BusinessException) WithMessage(msg string) *BusinessException {
	return &BusinessException{code: b.code, msg: msg}
}

// WithCode returns a new BusinessException with the updated code.
func (b *BusinessException) WithCode(code int) *BusinessException {
	return &BusinessException{code: code, msg: b.msg}
}

func NewBusinessErr(msg string) *BusinessException {
	return &BusinessException{code: DefaultBusinessCode, msg: msg}
}

// StackError represents an error with captured stack trace information.
// It implements fmt.Formatter to provide flexible output formatting:
//   - %s, %v: prints only the error message
//   - %+v: prints error message followed by full stack trace
//   - %q: prints quoted error message
//
// The stack trace is captured at the point where the error is created,
// making it useful for debugging and error tracking in production.
//
// Examples:
//
//	err := NewStackError("database connection failed")
//	fmt.Printf("%s\n", err)   // Output: database connection failed
//	fmt.Printf("%+v\n", err)  // Output: database connection failed
//	                          //         github.com/user/pkg.Function
//	                          //             /path/to/file.go:123
//	                          //         ...
type StackError struct {
	msg string // Error message describing the exception.
	st  string // Stack trace information.
}

// Error returns the error message without stack trace.
// This implements the error interface.
func (s *StackError) Error() string {
	return s.msg
}

// Format implements fmt.Formatter to support different output formats.
// Use %+v to include stack trace, other formats show only the message.
func (s *StackError) Format(f fmt.State, verb rune) {
	switch verb {
	case 'v':
		if f.Flag('+') {
			// %+v: print message and stack trace
			fmt.Fprint(f, s.msg)
			fmt.Fprint(f, "\n")
			fmt.Fprint(f, s.st)
			return
		}
		fallthrough
	case 's':
		// %s, %v: print only message
		fmt.Fprint(f, s.msg)
	case 'q':
		// %q: print quoted message
		fmt.Fprintf(f, "%q", s.msg)
	}
}

// NewStackError creates a new StackError with the given message and captures the current stack trace.
// The stack trace captures up to 8 frames starting from the caller's location.
//
// Example:
//
//	err := NewStackError("operation failed")
//	log.Printf("%+v", err) // Logs message with stack trace
func NewStackError(msg string) *StackError {
	stack := stacktrace.Capture(1, 8)
	defer stack.Free()
	return &StackError{msg, stack.ToString()}
}

// WrapWithStack wraps an existing error with stack trace information.
// Returns nil if the input error is nil.
// If the error is already a StackError, consider using it directly to preserve the original stack.
//
// Example:
//
//	if err := someOperation(); err != nil {
//	    return WrapWithStack(err)
//	}
func WrapWithStack(err error) *StackError {
	if err == nil {
		return nil
	}
	stack := stacktrace.Capture(1, 8)
	defer stack.Free()
	return &StackError{err.Error(), stack.ToString()}
}

// StackTrace returns the captured stack trace as a string.
// The format is compatible with standard Go stack trace output.
func (s *StackError) StackTrace() string {
	return s.st
}

// String returns the complete string representation including both
// the error message and stack trace, equivalent to fmt.Sprintf("%+v", err).
func (s *StackError) String() string {
	return fmt.Sprintf("%s\n%s", s.msg, s.st)
}
