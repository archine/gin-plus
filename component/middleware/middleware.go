package middleware

import (
	"fmt"
	"github.com/archine/gin-plus/v4/component/gplog"
	"github.com/archine/gin-plus/v4/component/stacktrace"
	"github.com/archine/gin-plus/v4/exception"
	"github.com/archine/gin-plus/v4/resp"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// Cors Cross-domain middleware
func Cors() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowMethods:     []string{"*"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return true
		},
	})
}

// GlobalExceptionInterceptor gin global exception interceptor
// add via gin middleware.
// thrown when the exception type is string and the BusinessException
func GlobalExceptionInterceptor(ctx *gin.Context) {
	defer func() {
		if r := recover(); r != nil {
			switch t := r.(type) {
			case *exception.BusinessException:
				resp.DirectRespWithCode(ctx, t.Code(), t.Error())
			case *exception.StackError:
				gplog.WithContext(ctx).Error(fmt.Sprintf("%s\n%s", t.Error(), t.StackTrace()))
				resp.ServerError(ctx, true)
			default:
				trace := getTrace()
				gplog.WithContext(ctx).Error(fmt.Sprintf("%v\n%s", r, trace))
				resp.ServerError(ctx, true)
			}
		}
	}()
	ctx.Next()
}

func getTrace() string {
	stack := stacktrace.Capture(4, 8)
	defer stack.Free()
	return stack.ToString()
}
