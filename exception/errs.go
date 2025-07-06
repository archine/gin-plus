package exception

import (
	"fmt"
	"github.com/archine/gin-plus/v4/exception/stacktrace"
)

// BusinessException represents a service-level exception that includes stack trace information.
// The default business error code is set to 10400, which corresponds to bcode.BadRequest.
//
// When returned via resp.DirectRespErr, this error is not treated as an unknown error,
// so resp.ServerError will not be triggered.
type BusinessException struct {
	code int    // Business error code, default is 10400.
	msg  string // Error message describing the exception.
}

func (b *BusinessException) Error() string {
	return b.msg
}

func (b *BusinessException) Code() int {
	return b.code
}

func NewBusinessErr(msg string) *BusinessException {
	return &BusinessException{code: 10400, msg: msg}
}

func NewBusinessErrWithCode(bcode int, msg string) *BusinessException {
	return &BusinessException{code: bcode, msg: msg}
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
type StackError struct {
	msg string // Error message describing the exception.
	st  string // Stack trace information.
}

func (s *StackError) Error() string {
	return s.msg
}

func NewStackErr(msg string) *StackError {
	stack := stacktrace.Capture(1, 8)
	defer stack.Free()
	return &StackError{msg, stack.ToString()}
}

// WithStack wraps an error with a stack trace.
func WithStack(err error) *StackError {
	if err == nil {
		return nil
	}
	stack := stacktrace.Capture(1, 8)
	defer stack.Free()
	return &StackError{err.Error(), stack.ToString()}
}

// StackTrace returns the stack trace of the error.
func (s *StackError) StackTrace() string {
	return s.st
}

// ToString returns the string representation of the StackError,
// which includes the error message and stack trace.
func (s *StackError) ToString() string {
	return fmt.Sprintf("%s\n%s", s.msg, s.st)
}
