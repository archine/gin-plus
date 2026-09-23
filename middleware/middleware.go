package middleware

import (
	"errors"
	"fmt"
	"net/http"
	"syscall"

	"github.com/archine/gin-plus/v4/component/gplog"
	"github.com/archine/gin-plus/v4/exception"
	"github.com/archine/gin-plus/v4/resp"
	"github.com/gin-gonic/gin"
)

// Recovery returns a middleware that recovers from panics, see RecoveryWithMessage.
func Recovery() gin.HandlerFunc {
	return RecoveryWithMessage("Internal Server Error")
}

// RecoveryWithMessage returns a middleware that recovers from panics and answers with
// DefaultSystemErrorCode and the given message.
//
// Behavior:
//   - the panic value is logged at Error level with the stack trace of the panicking code;
//   - http.ErrAbortHandler is re-panicked, it is the standard way to abort a response and
//     net/http handles it silently;
//   - a broken client connection is only logged, there is nobody left to answer;
//   - if the response has already been partially written, the chain is aborted without
//     writing again, which would corrupt the body.
//
// Register it before other middlewares so that it covers all of them.
func RecoveryWithMessage(message string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			r := recover()
			if r == nil {
				return
			}

			err, _ := r.(error)
			if errors.Is(err, http.ErrAbortHandler) {
				panic(r)
			}
			if errors.Is(err, syscall.EPIPE) || errors.Is(err, syscall.ECONNRESET) {
				gplog.WarnWithCtx(ctx, "Client connection lost: "+err.Error())
				ctx.Abort()
				return
			}

			// Skip this deferred function; the panic machinery frames above the failing line
			// are trimmed by Stack, however many the runtime uses.
			stack := exception.CaptureStackTrace(1, 16)
			gplog.ErrorWithCtx(ctx, fmt.Sprintf("Panic recovered: %v\n%s", r, stack.Full()))

			if ctx.Writer.Written() {
				ctx.Abort()
				return
			}
			resp.Code(ctx, exception.DefaultSystemErrorCode, message)
		}()
		ctx.Next()
	}
}
