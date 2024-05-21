package exception

import (
	"bytes"
	"github.com/archine/gin-plus/v3/plugin/logger"
	"runtime"
)

// BusinessException the service level exception, equivalent to resp.BadRequest
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

// OrThrow if err not nil, panic
func OrThrow(err error) {
	if err != nil {
		panic(err)
	}
}

// PrintStack print full stack
func PrintStack(err error) {
	var buf [2048]byte
	n := runtime.Stack(buf[:], false)
	logger.Log.Error("%s %s", err.Error(), string(buf[:n]))
}

// PrintSimpleStack Print short stack information
func PrintSimpleStack(err error) {
	var buf [2048]byte
	n := runtime.Stack(buf[:], false)
	lines := bytes.Split(buf[:n], []byte("\n"))
	logger.Log.Error("%s\n%s", err.Error(), string(bytes.Join(lines[7:11], []byte("\n"))))
}
