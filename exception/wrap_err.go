package exception

import (
	"errors"
	"fmt"
)

// WrapError adds business context on top of another error while preserving
// the original error chain for errors.Is and errors.As.
type WrapError struct {
	msg string // Current business message.
	err error  // Wrapped error.
}

// Error returns only the business message chain.
//
// This keeps user-facing messages concise. Use Full when the final wrapped
// error detail or stack trace is required.
func (w *WrapError) Error() string {
	var wrappedErr *WrapError
	if errors.As(w.err, &wrappedErr) {
		return w.msg + ": " + wrappedErr.Error()
	}
	return w.msg
}

// Full returns the full message chain plus the terminal wrapped error.
//
// If the leaf error supports fmt.Formatter, %+v is used so richer output such
// as stack traces is preserved.
func (w *WrapError) Full() string {
	return w.msg + ": " + fullError(w.err)
}

// Unwrap returns the wrapped error.
func (w *WrapError) Unwrap() error {
	return w.err
}

// Format implements fmt.Formatter.
//
// %s and %v print the business message chain.
// %+v prints the complete wrapped error.
// %q prints the quoted business message chain.
func (w *WrapError) Format(f fmt.State, verb rune) {
	switch verb {
	case 'v':
		if f.Flag('+') {
			fmt.Fprint(f, w.Full())
			return
		}
		fallthrough
	case 's':
		fmt.Fprint(f, w.Error())
	case 'q':
		fmt.Fprintf(f, "%q", w.Error())
	}
}

// WrapBusinessMsg wraps an error with a business-level message.
func WrapBusinessMsg(err error, msg string) error {
	if err == nil {
		return nil
	}

	return &WrapError{msg: msg, err: err}
}

// WrapBusinessMsgF wraps an error with a formatted business-level message.
func WrapBusinessMsgF(err error, format string, args ...any) error {
	if err == nil {
		return nil
	}

	return &WrapError{msg: fmt.Sprintf(format, args...), err: err}
}

func fullError(err error) string {
	if f, ok := err.(interface{ Full() string }); ok {
		return f.Full()
	}
	return err.Error()
}
