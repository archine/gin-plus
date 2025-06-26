package ctrl

import "github.com/gin-gonic/gin"

// MethodInterceptor allows for pre- and post-processing of API method calls.
type MethodInterceptor interface {
	// Predicate determines whether to intercept the request.
	Predicate(ctx *gin.Context) bool

	// PreHandle is triggered before the API method is invoked.
	// If you want to abort the current request, call ctx.Abort() and handle the response inside this method.
	PreHandle(ctx *gin.Context)

	// PostHandle is triggered after the API method is invoked.
	// If you want to abort the current request, call ctx.Abort() and handle the response inside this method.
	PostHandle(ctx *gin.Context)
}
