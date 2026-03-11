package exception

const (
	// DefaultBusinessCode is the default business error code, set to 10400.
	DefaultBusinessCode = 10400

	// DefaultNoLoginCode is the default no-login error code, set to 10401.
	DefaultNoLoginCode = 10401

	// DefaultTokenExpired is the default token expired error code, set to 10402.
	DefaultTokenExpired = 10402

	// DefaultForbiddenCode is the default forbidden error code, set to 10403.
	DefaultForbiddenCode = 10403

	// DefaultSystemErrorCode is the default system error code, set to 10500.
	DefaultSystemErrorCode = 10500
)

// BusinessException represents a business logic error with a specific code and message.
// When returned through resp.Error(), this type will be detected.
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

// WithCode returns a new BusinessException with the updated code.
func (b *BusinessException) WithCode(code int) *BusinessException {
	return &BusinessException{code: code, msg: b.msg}
}

func NewBusinessErr(msg string) *BusinessException {
	return &BusinessException{code: DefaultBusinessCode, msg: msg}
}
