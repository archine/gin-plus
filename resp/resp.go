package resp

import (
	"errors"
	"fmt"
	"github.com/archine/gin-plus/v3/exception"
	"github.com/archine/gin-plus/v3/internal"
	"github.com/archine/gin-plus/v3/module/constant/errs"
	"github.com/archine/gin-plus/v3/module/pool"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"net/http"
	"reflect"
	"strings"
)

/*
Package ResponseHandler provides a unified way to handle HTTP responses in a Gin-based web application.

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
	Total     int64       `json:"total"`      // Total count of items.
	PageSize  int         `json:"page_size"`  // Number of items per page.
	PageIndex int         `json:"page_index"` // Index of the current page.
	Data      interface{} `json:"data"`       // Data for the current page.
}

// Result represents a standard response structure.
type Result struct {
	ctx     *gin.Context
	Code    int         `json:"code"`               // Business code.
	TraceId string      `json:"trace_id,omitempty"` // Optional trace ID for tracking, can be empty.
	Message string      `json:"msg"`                // Business message.
	Data    interface{} `json:"ret,omitempty"`      // Response data, can be empty.
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
	r.TraceId = r.ctx.GetString("trace_id")
	if len(httpCode) > 0 {
		r.ctx.JSON(httpCode[0], r)
	} else {
		r.ctx.JSON(http.StatusOK, r)
	}
	r.ctx.Abort()

	// Reset fields and release the object back to the pool
	r.ctx = nil
	r.Code = 0
	r.Message = ""
	r.Data = nil
	r.TraceId = ""
	Free(r)
}

// InitResp initializes a new Resp object from the pool.
func InitResp(ctx *gin.Context) Resp {
	return _resultPool.Get().(Resp).WithContext(ctx)
}

// BadRequest handles business-related errors.
// Returns true if the condition is true.
func BadRequest(ctx *gin.Context, condition bool, msg ...string) bool {
	if condition {
		message := strings.Join(msg, ",")
		if message == "" {
			message = errs.BadRequestErr.Error()
		}
		InitResp(ctx).WithBasic(errs.BadRequestErr.Code(), message, nil).To()
	}
	return condition
}

// DirectBadRequest directly returns a business-related error.
func DirectBadRequest(ctx *gin.Context, format string, args ...any) {
	BadRequest(ctx, true, fmt.Sprintf(format, args...))
}

// ParamValidation performs parameter validation.
// Returns false if the validation fails, and sends an appropriate error message.
func ParamValidation(ctx *gin.Context, obj interface{}) bool {
	err := ctx.ShouldBind(obj)
	if err == nil {
		return true
	}
	msg := GetValidMsg(err, obj)
	if msg == "" {
		msg = errs.ParamErr.Error()
	}
	InitResp(ctx).WithBasic(errs.ParamErr.Code(), msg, nil).To()
	return false
}

// Forbidden handles requests where the server understands the request but refuses to execute it.
// Returns true if the condition is true.
func Forbidden(ctx *gin.Context, condition bool, msg ...string) bool {
	if condition {
		message := strings.Join(msg, ",")
		if message == "" {
			message = errs.ForbiddenErr.Error()
		}
		InitResp(ctx).WithBasic(errs.ForbiddenErr.Code(), message, nil).To()
	}
	return condition
}

// NoLogin handles situations where the user is not logged in.
// Returns true if the condition is true.
func NoLogin(ctx *gin.Context, condition bool, msg ...string) bool {
	if condition {
		message := strings.Join(msg, ",")
		if message == "" {
			message = errs.NoLoginErr.Error()
		}
		InitResp(ctx).WithBasic(errs.NoLoginErr.Code(), message, nil).To(http.StatusUnauthorized)
	}
	return condition
}

// LoginExpired handles cases where the user's login session has expired.
// Returns true if the condition is true.
func LoginExpired(ctx *gin.Context, condition bool, msg ...string) bool {
	if condition {
		message := strings.Join(msg, ",")
		if message == "" {
			message = errs.TokenExpiredErr.Error()
		}
		InitResp(ctx).WithBasic(errs.TokenExpiredErr.Code(), message, nil).To(http.StatusUnauthorized)
	}
	return condition
}

// Ok sends a standard success response with no data.
func Ok(ctx *gin.Context) {
	InitResp(ctx).WithBasic(0, "ok", nil).To()
}

// Json sends a standard success response with data.
func Json(ctx *gin.Context, data any) {
	InitResp(ctx).WithBasic(0, "ok", data).To()
}

// ServerError handles server exceptions.
// Returns true if the condition is true.
func ServerError(ctx *gin.Context, condition bool, msg ...string) bool {
	if condition {
		message := strings.Join(msg, ",")
		if message == "" {
			message = errs.ServerErr.Error()
		}
		InitResp(ctx).WithBasic(errs.ServerErr.Code(), message, nil).To()
	}
	return condition
}

// DirectRespWithCode directly responds with a custom business code.
func DirectRespWithCode(ctx *gin.Context, bCode int, format string, args ...any) {
	InitResp(ctx).WithBasic(bCode, fmt.Sprintf(format, args...), nil).To()
}

// DirectRespErr responds directly with any error.
//
// The function first checks if the error is of type exception.StackError.
// If it is, the stack trace is logged for debugging purposes, but the function
// continues to check for other error types.
//
// If the error is a BusinessException, the function responds with the corresponding
// business error code and message.
//
// If no specific error type is identified, a generic server error response is returned.
func DirectRespErr(ctx *gin.Context, err error) {
	var stackErr *exception.StackError
	if errors.As(err, &stackErr) {
		internal.Logger.Error("%s\n%s", err.Error(), stackErr.StackTrace())
	}
	var businessErr *exception.BusinessException
	if errors.As(err, &businessErr) {
		DirectRespWithCode(ctx, businessErr.Code(), businessErr.Error())
		return
	}
	ServerError(ctx, true)
}

// ChangeResultType changes the result type used by the result pool.
func ChangeResultType(f func() Resp) {
	_resultPool = pool.New(f)
}

// Free releases the result back to the pool.
func Free(resp Resp) {
	_resultPool.Put(resp)
}

// GetValidMsg extracts a user-friendly error message from a validation error.
// It retrieves the message associated with the field in the provided object,
// using the custom tag or a default message if not found.
//
// Parameters:
// - err: The error to extract a message from. It should be of type validator.ValidationErrors.
// - obj: The object used to retrieve field-specific messages from tags.
func GetValidMsg(err error, obj interface{}) string {
	if err == nil {
		return ""
	}
	if obj == nil {
		return err.Error() // Return the error message as is if obj is nil
	}
	var validationErrors validator.ValidationErrors
	if !errors.As(err, &validationErrors) {
		// If the error is not of type validator.ValidationErrors, return the default error message
		return err.Error()
	}

	objType := reflect.TypeOf(obj)
	if objType.Kind() == reflect.Ptr {
		objType = objType.Elem()
	}

	for _, ve := range validationErrors {
		if field, exists := objType.FieldByName(ve.Field()); exists {
			// Retrieve custom error message from tag
			if message := field.Tag.Get(ve.Tag() + "Msg"); message != "" {
				return message
			}
			// Retrieve default error message from tag
			if message := field.Tag.Get("msg"); message != "" {
				return message
			}
		}
	}
	return ""
}
