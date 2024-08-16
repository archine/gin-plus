package errs

type Error struct {
	code int
	msg  string
}

func (e *Error) Error() string {
	return e.msg
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
	BadRequestErr = &Error{40000, "操作失败"}

	// ParamErr is a parameter error
	ParamErr = &Error{40001, "参数错误"}

	// ForbiddenErr is a forbidden error
	ForbiddenErr = &Error{40300, "访问被拒绝"}

	// ServerErr is a server error
	ServerErr = &Error{50000, "服务器异常，请联系管理员"}

	// TokenExpiredErr is a token expired error
	TokenExpiredErr = &Error{40101, "登录已过期"}

	// UnauthorizedErr is an unauthorized error
	UnauthorizedErr = &Error{40100, "未授权"}

	// NoLoginErr is a no login error
	NoLoginErr = &Error{40102, "当前未登录"}
)
