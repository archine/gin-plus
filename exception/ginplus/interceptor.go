package ginplus

import (
	"fmt"
	"github.com/archine/gin-plus/v3/exception"
	"github.com/archine/gin-plus/v3/resp"
	"github.com/gin-gonic/gin"
)

// GlobalExceptionInterceptor gin global exception interceptor
// add via gin middleware.
// thrown when the exception type is string and the BusinessException
func GlobalExceptionInterceptor(context *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			switch t := r.(type) {
			case *exception.BusinessException:
				fmt.Printf("%+v", stacktrace.CallersWithSize(16))
				resp.DirectRespWithCode(context, t.Code, t.Msg)
			case *exception.StackBusinessError:
				fmt.Printf("%+v", t)
				resp.DirectRespWithCode(context, t.Code, t.Msg)
			default:
				fmt.Printf("%+v", stacktrace.Callers())
				resp.ServerError(context, true)
			}
		}
	}()
	context.Next()
}
