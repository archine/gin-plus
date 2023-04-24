package exception

import (
	"github.com/archine/gin-plus/v2/resp"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"runtime"
)

// BusinessException the service level exception, the service code is -10400, equivalent to resp.BadRequest
type BusinessException struct {
	message string
}

func (b BusinessException) Error() string {
	return b.message
}

func NewBusinessException(msg string) BusinessException {
	return BusinessException{message: msg}
}

func printStack(err error) {
	var buf [2048]byte
	n := runtime.Stack(buf[:], false)
	log.Errorf("%s %s", err.Error(), string(buf[:n]))
}

// GlobalExceptionInterceptor gin global exception interceptor
// Added via gin middleware
func GlobalExceptionInterceptor(context *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			if s, ok := r.(BusinessException); ok {
				resp.BadRequest(context, true, s.Error())
				return
			}
			if s, ok := r.(error); ok {
				printStack(s)
				resp.SeverError(context, s)
				return
			}
			log.Error(r)
			resp.InitResp(context, http.StatusOK).
				WithCode(resp.INTERNAL_SERVER_CODE).
				WithMessage(http.StatusText(http.StatusInternalServerError)).
				To()
		}
	}()
	context.Next()
}

// OrThrow if err not nil, panic
func OrThrow(err error) {
	if err != nil {
		panic(err)
	}
}
