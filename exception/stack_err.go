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
	err error  // Wrapped error (set via WrapWithStack).
	st  string // Captured stack trace.
}

// Error returns the error message without stack trace.
func (s *StackError) Error() string {
	if s == nil {
		return ""
	}
	return s.err.Error()
}

// Full returns the error detail followed by the captured stack trace.
//
// When the wrapped error supports Full (e.g. WrapError), its full
// representation is used so no information is lost.
func (s *StackError) Full() string {
	if s == nil {
		return ""
	}
	return s.err.Error() + "\n" + s.st
}

// StackTrace returns the captured stack trace.
func (s *StackError) StackTrace() string {
	if s == nil {
		return ""
	}
	return s.st
}

// Unwrap returns the wrapped error for errors.Is / errors.As.
func (s *StackError) Unwrap() error {
	if s == nil {
		return nil
	}
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

// WrapWithStack adds a stack trace to an existing error.
// Returns nil if err is nil.
// If err is already a *StackError, it is returned as-is to avoid redundant nesting.
func WrapWithStack(err error, skip ...int) error {
	if err == nil {
		return nil
	}
	var se *StackError
	if errors.As(err, &se) {
		return se
	}

	s := 1
	if len(skip) > 0 {
		s = skip[0]
	}
	st := CaptureStackTrace(s, 16)
	return &StackError{err: err, st: st.Full()}
}
