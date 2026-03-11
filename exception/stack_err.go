package exception

import (
	"errors"
	"fmt"
)

// StackError captures a stack trace alongside an error.
//
// Formatting verbs:
//   - %s, %v: error message only
//   - %+v: error message with full stack trace (same as Full)
//   - %q: quoted error message
type StackError struct {
	msg string // Direct message (set via NewStackError).
	err error  // Wrapped error (set via WrapWithStack).
	st  string // Captured stack trace.
}

// Error returns the error message without stack trace.
func (s *StackError) Error() string {
	if s.err != nil {
		return s.err.Error()
	}
	return s.msg
}

// Full returns the error detail followed by the captured stack trace.
//
// When the wrapped error supports Full (e.g. WrapError), its full
// representation is used so no information is lost.
func (s *StackError) Full() string {
	var detail string
	if s.err != nil {
		detail = fullError(s.err)
	} else {
		detail = s.msg
	}
	return detail + "\n" + s.st
}

// StackTrace returns the captured stack trace.
func (s *StackError) StackTrace() string {
	return s.st
}

// Unwrap returns the wrapped error for errors.Is / errors.As.
func (s *StackError) Unwrap() error {
	return s.err
}

// Format implements fmt.Formatter.
func (s *StackError) Format(f fmt.State, verb rune) {
	switch verb {
	case 'v':
		if f.Flag('+') {
			fmt.Fprint(f, s.Full())
			return
		}
		fallthrough
	case 's':
		fmt.Fprint(f, s.Error())
	case 'q':
		fmt.Fprintf(f, "%q", s.Error())
	}
}

// NewStackError creates a StackError with a message and captures the current stack trace.
// The trace captures up to 16 frames starting from the caller's location.
func NewStackError(msg string) *StackError {
	st := CaptureStackTrace(1, 16)
	return &StackError{msg: msg, st: st.Full()}
}

// WrapWithStack adds a stack trace to an existing error.
// Returns nil if err is nil.
// If err is already a *StackError, it is returned as-is to avoid redundant nesting.
func WrapWithStack(err error) *StackError {
	if err == nil {
		return nil
	}
	var se *StackError
	if errors.As(err, &se) {
		return se
	}
	st := CaptureStackTrace(1, 16)
	return &StackError{err: err, st: st.Full()}
}
