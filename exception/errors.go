package exception

import (
	"github.com/archine/gin-plus/v3/exception/stacktrace"
	"github.com/archine/gin-plus/v3/resp/bcode"
)

// BusinessException the service level exception.
// Procedure When returned via resp.DirectRespErr, this error is not interpreted as an unknown error and resp.ServerError is raised
type BusinessException struct {
	Code int
	Msg  string
}

func (b *BusinessException) Error() string {
	return b.Msg
}

func NewBusinessErr(msg string) *BusinessException {
	return &BusinessException{40000, msg}
}

func NewBusinessErrWithCode(code int, msg string) *BusinessException {
	return &BusinessException{code, msg}
}

// StackBusinessError Service level exception with stack.
// Procedure When using fmt.Printf, stack information is printed when format is '%+v', and only error messages are printed otherwise.
//
// # When returned via resp.DirectRespErr, this error is not interpreted as an unknown error and resp.ServerError is raised
//
// Usage:
//
//	err := NewBusinessErrWithStack("error message")
//	fmt.Printf("%+v", err)
type StackBusinessError struct {
	Code int
	Msg  string
	s    *stacktrace.Stack
}

func (s *StackBusinessError) Error() string {
	return s.Msg
}

func NewStackBusinessErr(msg string) *StackBusinessError {
	return &StackBusinessError{bcode.BadRequest, msg, stacktrace.Callers()}
}

func NewStackBusinessErrWithCode(code int, msg string) *StackBusinessError {
	return &StackBusinessError{code, msg, stacktrace.Callers()}
}

// OrThrow if err not nil, panic
func OrThrow(err error) {
	if err != nil {
		panic(err)
	}
}
