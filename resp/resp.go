package resp

import (
	"errors"
	"fmt"
	"net/http"

	"github.com/archine/gin-plus/v4/component/gplog"
	"github.com/archine/gin-plus/v4/component/pool"
	"github.com/archine/gin-plus/v4/exception"
	"github.com/gin-gonic/gin"
)

/*
Package ResponseHandler provides a unified way to handle HTTP responses in a Gin-based web app.

This package includes utilities for managing standard and error responses, along with a result pooling mechanism to optimize performance.

Usage:
- Integrate this package into Gin controllers for consistent response formatting and efficient object reuse.
*/

// ResultPool is a pool that manages reusable Result objects.
var _resultPool = pool.New(func() Resp {
	return &Result{}
})

// Resp defines the interface for responses to be sent to the client.
type Resp interface {
	// WithBasic sets the basic properties of the response.
	WithBasic(businessCode int, msg string, data any) Resp

	// WithContext sets the context, which is mandatory to avoid null pointer exceptions.
	WithContext(ctx *gin.Context) Resp

	// To sends the response to the client. The optional httpCode parameter allows specifying an HTTP status code.
	To(httpCode ...int)
}

// PaginationResult represents a paginated response.
type PaginationResult struct {
	Total     int64 `json:"total"`      // Total count of items.
	PageSize  int   `json:"page_size"`  // Number of items per page.
	PageIndex int   `json:"page_index"` // Index of the current page.
	Data      any   `json:"data"`       // Data for the current page.
}

// Result represents a standard response structure.
type Result struct {
	ctx     *gin.Context
	Code    int    `json:"code"`           // Business code.
	Message string `json:"msg"`            // Business message.
	Data    any    `json:"data,omitempty"` // Response data, can be empty.
}

// WithBasic sets the basic properties of the Result.
func (r *Result) WithBasic(code int, msg string, data any) Resp {
	r.Code = code
	r.Message = msg
	r.Data = data
	return r
}

// WithContext sets the Gin context for the Result.
func (r *Result) WithContext(ctx *gin.Context) Resp {
	r.ctx = ctx
	return r
}

// To sends the Result as a JSON response to the client and releases the object back to the pool.
func (r *Result) To(httpCode ...int) {
	if r.ctx == nil {
		panic("Response context is nil")
	}

	if len(httpCode) > 0 {
		r.ctx.JSON(httpCode[0], r)
	} else {
		r.ctx.JSON(http.StatusOK, r)
	}

	r.ctx.Abort()
	r.ctx = nil
	r.Code = 0
	r.Message = ""
	r.Data = nil

	_resultPool.Put(r)
}

// InitResp initializes a new Resp object from the pool.
func InitResp(ctx *gin.Context) Resp {
	return _resultPool.Get().(Resp).WithContext(ctx)
}

// BadRequest returns a business-related error.
func BadRequest(ctx *gin.Context, format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	if message == "" {
		message = http.StatusText(http.StatusBadRequest)
	}
	InitResp(ctx).WithBasic(exception.DefaultBusinessCode, message, nil).To()
}

// ParamValidation performs parameter validation.
// Returns false if the validation fails, and sends an appropriate error message.
func ParamValidation(ctx *gin.Context, obj any) bool {
	err := ctx.ShouldBind(obj)
	if err == nil {
		return true
	}
	InitResp(ctx).WithBasic(exception.DefaultBusinessCode, "Invalid parameters", nil).To()
	return false
}

// Forbidden handles situations where access is forbidden.
func Forbidden(ctx *gin.Context, format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	if message == "" {
		message = http.StatusText(http.StatusForbidden)
	}
	InitResp(ctx).WithBasic(exception.DefaultForbiddenCode, message, nil).To(http.StatusForbidden)
}

// NoLogin handles situations where the user is not logged in.
func NoLogin(ctx *gin.Context, format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	if message == "" {
		message = "Not logged in"
	}
	InitResp(ctx).WithBasic(exception.DefaultNoLoginCode, message, nil).To(http.StatusUnauthorized)
}

// LoginExpired handles cases where the user's login session has expired.
func LoginExpired(ctx *gin.Context, format string, args ...any) {
	message := fmt.Sprintf(format, args...)
	if message == "" {
		message = "Login expired"
	}
	InitResp(ctx).WithBasic(exception.DefaultTokenExpired, message, nil).To(http.StatusUnauthorized)
}

// Ok sends a standard success response with no data.
func Ok(ctx *gin.Context) {
	InitResp(ctx).WithBasic(0, "ok", nil).To()
}

// Json sends a standard success response with data.
func Json(ctx *gin.Context, data any) {
	InitResp(ctx).WithBasic(0, "ok", data).To()
}

// Code responds with a custom business code and message.
func Code(ctx *gin.Context, code int, format string, args ...any) {
    InitResp(ctx).WithBasic(code, fmt.Sprintf(format, args...), nil).To()
}

// Error responds with an error
//
// The function first checks if the error is of type exception.StackError.
// If it is, the stack trace is logged for debugging purposes, but the function
// continues to check for other error types.
//
// If the error is a BusinessException, the function responds with the corresponding
// business error code and message.
//
// If no specific error type is identified, a generic server error response is returned.
func Error(ctx *gin.Context, err error) {
	if err == nil {
		return
	}
	var stackErr *exception.StackError
	if errors.As(err, &stackErr) {
		gplog.ErrorWithCtx(ctx, stackErr.ToString())
	}

	var businessErr *exception.BusinessException
	if errors.As(err, &businessErr) {
		InitResp(ctx).WithBasic(businessErr.Code(), businessErr.Error(), nil).To()
		return
	}

	gplog.ErrorWithCtx(ctx, fmt.Sprintf("Internal Server Error: %v", err))
	InitResp(ctx).WithBasic(exception.DefaultSystemErrorCode, "Internal Server Error", nil).To()
}
