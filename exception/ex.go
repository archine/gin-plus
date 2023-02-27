package exception

import (
	"github.com/archine/gin-plus/v2/resp"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"runtime"
)

func printStack() {
	var buf [2048]byte
	n := runtime.Stack(buf[:], false)
	log.Errorf("[painc] %s\n", string(buf[:n]))
}

// GlobalExceptionInterceptor gin global exception interceptor
// Added via gin middleware
func GlobalExceptionInterceptor(context *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			if s, ok := r.(error); ok {
				printStack()
				resp.SeverError(context, s)
				return
			}
			log.Error(r)
			resp.InitResp(context, http.StatusOK).WithCode(resp.InternalServerError).WithMessage(resp.CodeMsgMap[resp.InternalServerError]).To()
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
