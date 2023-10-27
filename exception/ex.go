package exception

import (
	"bytes"
	"github.com/archine/gin-plus/v3/resp"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"runtime"
)

// BusinessException the service level exception, the service code is -10400, equivalent to resp.BadRequest
type BusinessException struct {
	error
}

type SysException struct {
	error
}

func printStack(err error) {
	var buf [2048]byte
	n := runtime.Stack(buf[:], false)
	log.Errorf("%s %s", err.Error(), string(buf[:n]))
}

func printSimpleStack(err error) {
	var buf [2048]byte
	n := runtime.Stack(buf[:], false)
	lines := bytes.Split(buf[:n], []byte("\n"))
	log.Errorf("%s\n%s", err.Error(), string(bytes.Join(lines[9:11], []byte("\n"))))
}

// GlobalExceptionInterceptor gin global exception interceptor
// add via gin middleware.
// an error of -10400 is thrown when the exception type is string and the BusinessException
func GlobalExceptionInterceptor(context *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			switch t := r.(type) {
			case string:
				resp.BadRequest(context, true, t)
			case BusinessException:
				printSimpleStack(t)
				resp.BadRequest(context, true, t.Error())
			case SysException:
				printSimpleStack(t)
				resp.SeverError(context, true)
			case error:
				printStack(t)
				resp.SeverError(context, true)
			default:
				log.Error(t)
				resp.SeverError(context, true)
			}
			return
		}
	}()
	context.Next()
}

// OrThrow if err not nil, panic
func OrThrow(err error) {
	if err != nil {
		panic(SysException{err})
	}
}

// OrThrowBusiness if err not nil, panic
func OrThrowBusiness(err error) {
	if err != nil {
		panic(BusinessException{err})
	}
}
