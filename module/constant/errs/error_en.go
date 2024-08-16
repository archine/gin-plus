// Description: Error constants in English.
//go:build en
// +build en

package errs

type Error struct {
	code int
	msg  string
}

func (e *Error) Error() string {
	return e.Msg
}

// Code returns the error code
func (e *Error) Code() int {
	return e.code
}

// NewError creates a new error
func NewError(code int, msg string) *Error {
	return &Error{code, msg}
}

var (
	// BadRequestErr is a bad request error
	BadRequestErr = &Error{40000, "Operation failed"}

	// ParamErr is a parameter error
	ParamErr = &Error{40001, "Parameter error"}

	// ForbiddenErr is a forbidden error
	ForbiddenErr = &Error{40300, "Access denied"}

	// ServerErr is a server error
	ServerErr = &Error{50000, "Server exception, please contact the administrator"}

	// TokenExpiredErr is a token expired error
	TokenExpiredErr = &Error{40101, "Login expired"}

	// UnauthorizedErr is an unauthorized error
	UnauthorizedErr = &Error{40100, "Unauthorized"}

	// NoLoginErr is a no login error
	NoLoginErr = &Error{40102, "Not logged in"}
)
