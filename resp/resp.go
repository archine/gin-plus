package resp

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
	"net/http"
	"reflect"
)

// Respond to the client assistant and return quickly

const (
	BAD_REQUEST_CODE      = -10400
	NONE_LOGIN_CODE       = -10401
	LOGIN_TOKEN_EXPIRED   = -10402
	NO_PERMISSION_CODE    = -10403
	TOO_MANY_REQUEST_CODE = -10429
	INTERNAL_SERVER_CODE  = -10500
	PARAM_FAILD_CODE      = -10600
)

type Resp interface {
	// WithMessage set the business message
	WithMessage(message string) Resp
	// WithCode set the business status code
	WithCode(code int) Resp
	// WithData set the response data
	WithData(data interface{}) Resp
	// To response client
	To()
}

// PaginationResult  Paging result
type PaginationResult struct {
	Total     int64 `json:"total"`      // Total count
	PageSize  int   `json:"page_size"`  // Page size
	PageIndex int   `json:"page_index"` // Current page index
	Data      any   `json:"data"`       // Response data
}

// Result Return result
type Result struct {
	ctx      *gin.Context `json:"-"`
	httpCode int          `json:"-"`             // http code
	Code     int          `json:"err_code"`      // business code
	Message  string       `json:"err_msg"`       // business message
	Data     any          `json:"ret,omitempty"` // Response data
}

func (r *Result) WithMessage(message string) Resp {
	r.Message = message
	return r
}

func (r *Result) WithCode(code int) Resp {
	r.Code = code
	return r
}

func (r *Result) WithData(data any) Resp {
	r.Data = data
	return r
}

func (r *Result) To() {
	r.ctx.Set("bcode", r.Code)
	r.ctx.JSON(r.httpCode, r)
}

// InitResp initialize a custom structure
func InitResp(ctx *gin.Context, httpCode int) *Result {
	return &Result{
		ctx:      ctx,
		httpCode: httpCode,
		Code:     0,
		Message:  http.StatusText(httpCode),
	}
}

// BadRequest business-related error returned.
// Return true means the condition is true
func BadRequest(ctx *gin.Context, condition bool, msg ...string) bool {
	if condition {
		message := "操作失败"
		if len(msg) > 0 {
			message = msg[0]
		}
		InitResp(ctx, http.StatusOK).WithCode(BAD_REQUEST_CODE).WithMessage(message).To()
	}
	return condition
}

// ParamInvalid invalid parameter.
//
//	Return true means the condition is true
func ParamInvalid(ctx *gin.Context, condition bool, msg ...string) bool {
	if condition {
		message := "参数无效"
		if len(msg) > 0 {
			message = msg[0]
		}
		InitResp(ctx, http.StatusOK).WithCode(PARAM_FAILD_CODE).WithMessage(message).To()
	}
	return condition
}

// Deprecated: please use ParamValidation. it will be removed in the future
func ParamValid(ctx *gin.Context, err error, obj interface{}) bool {
	if err == nil {
		return false
	}
	InitResp(ctx, http.StatusOK).WithCode(PARAM_FAILD_CODE).WithMessage(getValidMsg(err, obj)).To()
	return true
}

// ParamValidation parameter validation, return false means that the validation failed
func ParamValidation(ctx *gin.Context, obj interface{}) bool {
	err := ctx.ShouldBind(obj)
	if err == nil {
		return true
	}
	InitResp(ctx, http.StatusOK).WithCode(PARAM_FAILD_CODE).WithMessage(getValidMsg(err, obj)).To()
	return false
}

// NoPermission Insufficient permission error.
// Return true means the condition is true
func NoPermission(ctx *gin.Context, condition bool, msg ...string) bool {
	if condition {
		message := "权限不足"
		if len(msg) > 0 {
			message = msg[0]
		}
		InitResp(ctx, http.StatusOK).WithCode(NO_PERMISSION_CODE).WithMessage(message).To()
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
		InitResp(ctx, http.StatusUnauthorized).WithCode(NONE_LOGIN_CODE).WithMessage(message).To()
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
		InitResp(ctx, http.StatusUnauthorized).WithCode(LOGIN_TOKEN_EXPIRED).WithMessage(message).To()
	}
	return condition
}

// Ok Normal request with no data returned
func Ok(ctx *gin.Context) {
	InitResp(ctx, http.StatusOK).To()
}

// Json Normal request with data returned
func Json(ctx *gin.Context, data interface{}) {
	InitResp(ctx, http.StatusOK).WithCode(0).WithData(data).To()
}

// SeverError Server exception
// Return true means the condition is true
func SeverError(ctx *gin.Context, condition bool, msg ...string) bool {
	if condition {
		message := "服务器异常,请联系管理员!"
		if len(msg) > 0 {
			message = msg[0]
		}
		InitResp(ctx, http.StatusOK).WithCode(INTERNAL_SERVER_CODE).WithMessage(message).To()
	}
	return condition
}

// Custom User defined business code and message
func Custom(ctx *gin.Context, bCode int, msg string) {
	InitResp(ctx, http.StatusOK).WithCode(bCode).WithMessage(msg).To()
}

func getValidMsg(err error, obj interface{}) string {
	if obj == nil {
		return err.Error()
	}
	if errs, ok := err.(validator.ValidationErrors); ok {
		log.Error(err)
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
	log.Error(err.Error())
	return "参数无效"
}
