package resp

import (
	"errors"
	"fmt"
	"github.com/archine/gin-plus/v3/exception"
	"github.com/archine/gin-plus/v3/plugin/logger"
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	"net/http"
	"reflect"
	"sync"
)

// Respond to the client assistant and return quickly

const (
	BadRequestCode      = 40000
	NonLoginCode        = 40001
	TokenExpiredCode    = 40002
	ForbiddenCode       = 40003
	ParamValidationCode = 40010
	SystemErrorCode     = 50000
)

// ResultPool result pool
var resultPool = sync.Pool{
	New: func() interface{} {
		return &Result{}
	},
}

type Resp interface {
	// WithBasic Set the basic properties
	WithBasic(businessCode int, msg string, data any) Resp
	// WithContext Must be set the context, otherwise it will null pointer exception
	WithContext(ctx *gin.Context) Resp
	// To response client
	To(httpCode ...int)
}

// PaginationResult  Paging result
type PaginationResult struct {
	Total     int64       `json:"total"`      // Total count
	PageSize  int         `json:"page_size"`  // Page size
	PageIndex int         `json:"page_index"` // Current page index
	Data      interface{} `json:"data"`       // Response data
}

// Result Return result
type Result struct {
	ctx     *gin.Context
	Code    int         `json:"code"`               // business code
	TraceId string      `json:"trace_id,omitempty"` // trace id, optional, can be empty. you can manually set it.
	Message string      `json:"msg"`                // business message
	Data    interface{} `json:"ret,omitempty"`      // Response data
}

func (r *Result) WithBasic(code int, msg string, data any) Resp {
	r.Code = code
	r.Message = msg
	r.Data = data
	return r
}

func (r *Result) WithContext(ctx *gin.Context) Resp {
	r.ctx = ctx
	return r
}

func (r *Result) To(httpCode ...int) {
	r.TraceId = r.ctx.GetString("trace_id")
	r.ctx.Set("bcode", r.Code)
	if len(httpCode) > 0 {
		r.ctx.JSON(httpCode[0], r)
	} else {
		r.ctx.JSON(http.StatusOK, r)
	}
	// release
	r.ctx = nil
	r.Code = 0
	r.Message = ""
	r.Data = nil
	r.TraceId = ""
	Recycle(r)
}

// InitResp initialize a custom structure
func InitResp(ctx *gin.Context) Resp {
	return resultPool.Get().(Resp).WithContext(ctx)
}

// BadRequest business-related error returned.
// Return true means the condition is true
func BadRequest(ctx *gin.Context, condition bool, msg ...string) bool {
	if condition {
		message := "操作失败"
		if len(msg) > 0 {
			message = msg[0]
		}
		InitResp(ctx).WithBasic(BadRequestCode, message, nil).To()
	}
	return condition
}

// DirectBadRequest Directly return business-related errors.
func DirectBadRequest(ctx *gin.Context, format string, args ...any) {
	InitResp(ctx).WithBasic(BadRequestCode, fmt.Sprintf(format, args...), nil).To()
}

// ParamInvalid invalid parameter.
// Return true means the condition is true
func ParamInvalid(ctx *gin.Context, condition bool, msg ...string) bool {
	if condition {
		message := "参数错误"
		if len(msg) > 0 {
			message = msg[0]
		}
		InitResp(ctx).WithBasic(ParamValidationCode, message, nil).To()
	}
	return condition
}

// ParamValidation parameter validation, return false means that the validation failed
func ParamValidation(ctx *gin.Context, obj interface{}) bool {
	err := ctx.ShouldBind(obj)
	if err == nil {
		return true
	}
	InitResp(ctx).WithBasic(ParamValidationCode, getValidMsg(err, obj), nil).To()
	return false
}

// Forbidden Insufficient permission error.
// Return true means the condition is true
func Forbidden(ctx *gin.Context, condition bool, msg ...string) bool {
	if condition {
		message := "权限不足"
		if len(msg) > 0 {
			message = msg[0]
		}
		InitResp(ctx).WithBasic(ForbiddenCode, message, nil).To()
	}
	return condition
}

// NoLogin Not logged in.
// Return true means the condition is true
func NoLogin(ctx *gin.Context, condition bool, msg ...string) bool {
	if condition {
		message := "当前未登录"
		if len(msg) > 0 {
			message = msg[0]
		}
		InitResp(ctx).WithBasic(NonLoginCode, message, nil).To(http.StatusUnauthorized)
	}
	return condition
}

// LoginExpired Login expired
// Return true means the condition is true
func LoginExpired(ctx *gin.Context, condition bool, msg ...string) bool {
	if condition {
		message := "Token已过期"
		if len(msg) > 0 {
			message = msg[0]
		}
		InitResp(ctx).WithBasic(TokenExpiredCode, message, nil).To(http.StatusUnauthorized)
	}
	return condition
}

// Ok Normal request with no data returned
func Ok(ctx *gin.Context) {
	InitResp(ctx).WithBasic(0, "ok", nil).To()
}

// Json Normal request with data returned
func Json(ctx *gin.Context, data interface{}) {
	InitResp(ctx).WithBasic(0, "ok", data).To()
}

// SeverError Server exception
// Return true means the condition is true
func SeverError(ctx *gin.Context, condition bool, msg ...string) bool {
	if condition {
		message := "服务器异常,请联系管理员!"
		if len(msg) > 0 {
			message = msg[0]
		}
		InitResp(ctx).WithBasic(SystemErrorCode, message, nil).To()
	}
	return condition
}

// DirectRespWithCode Respond directly and customize the business code
func DirectRespWithCode(ctx *gin.Context, bCode int, format string, args ...any) {
	InitResp(ctx).WithBasic(bCode, fmt.Sprintf(format, args...), nil).To()
}

// DirectRespErr Respond directly with any err
func DirectRespErr(ctx *gin.Context, err error) {
	var businessErr *exception.BusinessException
	if errors.As(err, &businessErr) {
		DirectRespWithCode(ctx, businessErr.Code, businessErr.Msg)
		return
	}
	SeverError(ctx, true)
	exception.PrintStack(err)
}

// ChangeResultType Change the result type
func ChangeResultType(f func() Resp) {
	resultPool = sync.Pool{
		New: func() interface{} {
			return f()
		},
	}
}

// Recycle the result object
func Recycle(resp Resp) {
	resultPool.Put(resp)
}

func getValidMsg(err error, obj interface{}) string {
	if obj == nil {
		return err.Error()
	}
	var errs validator.ValidationErrors
	if errors.As(err, &errs) {
		logger.Log.Error(err.Error())
		getObj := reflect.TypeOf(obj)
		if getObj.Kind() == reflect.Ptr {
			getObj = getObj.Elem()
		}
		for _, e := range errs {
			if f, exist := getObj.FieldByName(e.Field()); exist {
				message := f.Tag.Get(e.Tag() + "Msg")
				if message == "" {
					message = f.Tag.Get("msg")
					if message == "" {
						return e.Error()
					}
				}
				return message
			}
		}
	}
	logger.Log.Error(err.Error())
	return "参数错误"
}
