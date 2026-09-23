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
	return fullError(s.err) + "\n" + s.st
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
// The trace starts at the caller of WrapWithStack; skip optionally drops that many more
// frames, for helpers wrapping errors on behalf of their own callers.
// Returns nil if err is nil.
// If err already carries a *StackError, it is returned unchanged: the innermost trace is the
// most precise one, and re-wrapping would only duplicate it.
func WrapWithStack(err error, skip ...int) error {
	if err == nil {
		return nil
	}
	if _, ok := errors.AsType[*StackError](err); ok {
		return err
	}

	s := 0
	if len(skip) > 0 {
		s = skip[0]
	}
	// +1 skips WrapWithStack itself.
	return &StackError{err: err, st: CaptureStackTrace(1+s, 16).Full()}
}
