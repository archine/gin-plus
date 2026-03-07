package middleware

import (
	"fmt"

	"github.com/archine/gin-plus/v4/component/gplog"
	"github.com/archine/gin-plus/v4/exception"
	"github.com/archine/gin-plus/v4/exception/stacktrace"
	"github.com/archine/gin-plus/v4/resp"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
)

// CORS returns a middleware that handles Cross-Origin Resource Sharing (CORS).
// It allows all origins, common HTTP methods, and credentials by default.
//
// Default configuration:
//   - Allows all origins (AllowOriginFunc returns true)
//   - Allows methods: GET, POST, PUT, PATCH, DELETE, HEAD, OPTIONS
//   - Allows all headers
//   - Exposes headers: Content-Length, Content-Type, Request-Id, X-Request-Id
//   - Allows credentials (cookies, authorization headers)
//
// Example:
//
//	app := gin.New()
//	app.Use(middleware.CORS())
func CORS() gin.HandlerFunc {
	return cors.New(cors.Config{
		AllowMethods:     []string{"GET", "POST", "PUT", "PATCH", "DELETE", "HEAD", "OPTIONS"},
		AllowHeaders:     []string{"*"},
		ExposeHeaders:    []string{"Content-Length", "Content-Type", "Request-Id", "X-Request-Id"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return true
		},
	})
}

// Recovery returns a middleware that recovers from panics and logs the error with stack trace.
// It catches all panics in the request handling chain and returns a standardized error response.
//
// Behavior:
//   - Catches any panic that occurs during request processing
//   - Returns HTTP 500 with error code DefaultSystemErrorCode
//   - Logs the panic value and full stack trace for debugging
//   - Prevents the application from crashing
//
// Example:
//
//	app := gin.New()
//	app.Use(middleware.Recovery())
//
// Note: This middleware should typically be registered first to catch panics
// from all subsequent middlewares and handlers.
func Recovery() gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// Return error response to client
				resp.Code(ctx, exception.DefaultSystemErrorCode, "Internal Server Error")

				// Log panic with stack trace for debugging
				stack := stacktrace.Capture(3, 16) // Capture up to 16 frames
				defer stack.Free()

				gplog.ErrorWithCtx(ctx, fmt.Sprintf("Panic recovered: %v\n%s", r, stack.ToString()))
			}
		}()
		ctx.Next()
	}
}

// RecoveryWithMessage returns a middleware that recovers from panics with a custom error message.
// Similar to Recovery() but allows customizing the error message returned to clients.
//
// Parameters:
//   - message: Custom error message to return to the client
//
// Example:
//
//	app := gin.New()
//	app.Use(middleware.RecoveryWithMessage("Service temporarily unavailable"))
func RecoveryWithMessage(message string) gin.HandlerFunc {
	return func(ctx *gin.Context) {
		defer func() {
			if r := recover(); r != nil {
				// Return custom error response to client
				resp.Code(ctx, exception.DefaultSystemErrorCode, "%s", message)

				// Log panic with stack trace for debugging
				stack := stacktrace.Capture(3, 16)
				defer stack.Free()

				gplog.ErrorWithCtx(ctx, fmt.Sprintf("Panic recovered: %v\n%s", r, stack.ToString()))
			}
		}()
		ctx.Next()
	}
}
