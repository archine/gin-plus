package interceptor

import (
	"github.com/archine/gin-plus/v3/exception"
	"github.com/archine/gin-plus/v3/plugin/logger"
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
				exception.PrintSimpleStack(t)
				resp.DirectRespWithCode(context, t.Code, t.Msg)
			case error:
				exception.PrintStack(t)
				resp.SeverError(context, true)
			default:
				logger.Log.Error("Unknown error: %v", r)
				resp.SeverError(context, true)
			}
		}
	}()
	context.Next()
}
