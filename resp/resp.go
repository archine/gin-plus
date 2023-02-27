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
	Success             = 0
	InternalServerError = -10500
	BadRequestError     = -10400
	TooManyRequestError = -10429
	NoPermissionError   = -10401
	NoneLoginError      = -10600
	DataExistsError     = -10601
	ParamValidError     = -10602
	TokenExpired        = -10603
)

var CodeMsgMap = map[int]string{
	Success:             "ok",
	InternalServerError: "系统出错,请联系管理员!",
	BadRequestError:     "操作失败",
	TooManyRequestError: "请求频繁,请稍后再试!",
	NoPermissionError:   "权限不足",
	NoneLoginError:      "当前未登录",
	DataExistsError:     "数据已存在",
	ParamValidError:     "参数错误",
	TokenExpired:        "token已过期",
}

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
	Total     int64       `json:"total"`      // Total count
	PageSize  int         `json:"page_size"`  // Page size
	PageIndex int         `json:"page_index"` // Current page index
	Data      interface{} `json:"data"`       // Response data
}

// Result Return result
type Result struct {
	ctx      *gin.Context `json:"-"`
	httpCode int          `json:"-"`             // http code
	Code     int          `json:"err_code"`      // business code
	Message  string       `json:"err_msg"`       // business message
	Data     interface{}  `json:"ret,omitempty"` // Response data
}

func (r *Result) WithMessage(message string) Resp {
	r.Message = message
	return r
}

func (r *Result) WithCode(code int) Resp {
	r.Code = code
	return r
}

func (r *Result) WithData(data interface{}) Resp {
	r.Data = data
	return r
}

func (r *Result) To() {
	r.ctx.Set("businessCode", r.Code)
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
// true means that flag is satisfied
func BadRequest(ctx *gin.Context, flag bool, msg ...string) bool {
	if flag {
		message := CodeMsgMap[BadRequestError]
		if len(msg) > 0 {
			message = msg[0]
		}
		InitResp(ctx, http.StatusOK).WithCode(BadRequestError).WithMessage(message).To()
	}
	return flag
}

// DataExists the data already exists, which is usually used when adding data.
// true means that flag is satisfied
func DataExists(ctx *gin.Context, flag bool, msg ...string) bool {
	if flag {
		message := CodeMsgMap[DataExistsError]
		if len(msg) > 0 {
			message = msg[0]
		}
		InitResp(ctx, http.StatusOK).WithCode(DataExistsError).WithMessage(message).To()
	}
	return flag
}

// ParamInvalid parameter check type error.
// true means that flag is satisfied
func ParamInvalid(ctx *gin.Context, flag bool, msg ...string) bool {
	if flag {
		message := CodeMsgMap[ParamValidError]
		if len(msg) > 0 {
			message = msg[0]
		}
		InitResp(ctx, http.StatusOK).WithCode(ParamValidError).WithMessage(message).To()
	}
	return flag
}

// ParamValid structure parameter validation type error.
// true means that flag is satisfied
func ParamValid(ctx *gin.Context, err error, obj interface{}) bool {
	if err == nil {
		return false
	}
	InitResp(ctx, http.StatusOK).WithCode(ParamValidError).WithMessage(getValidMsg(err, obj)).To()
	return true
}

// NoPermission Insufficient permission error.
// true means that flag is satisfied
func NoPermission(ctx *gin.Context, flag bool, msg ...string) bool {
	if flag {
		message := CodeMsgMap[NoPermissionError]
		if len(msg) > 0 {
			message = msg[0]
		}
		InitResp(ctx, http.StatusOK).WithCode(NoPermissionError).WithMessage(message).To()
	}
	return flag
}

// NoLogin Not logged in.
// true means that flag is satisfied
func NoLogin(ctx *gin.Context, flag bool, msg ...string) bool {
	if flag {
		message := CodeMsgMap[NoneLoginError]
		if len(msg) > 0 {
			message = msg[0]
		}
		InitResp(ctx, http.StatusOK).WithCode(NoneLoginError).WithMessage(message).To()
	}
	return flag
}

// LoginExpired Login expired
// true means that flag is satisfied
func LoginExpired(ctx *gin.Context, flag bool, msg ...string) bool {
	if flag {
		message := CodeMsgMap[TokenExpired]
		if len(msg) > 0 {
			message = msg[0]
		}
		InitResp(ctx, http.StatusOK).WithCode(TokenExpired).WithMessage(message).To()
	}
	return flag
}

// Ok Normal request returned, no data
func Ok(ctx *gin.Context) {
	InitResp(ctx, http.StatusOK).WithCode(0).To()
}

// Json Normal request returned with data
func Json(ctx *gin.Context, data interface{}) {
	InitResp(ctx, http.StatusOK).WithCode(0).WithData(data).To()
}

// SeverError Server level exception
func SeverError(ctx *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	log.Error(err.Error())
	InitResp(ctx, http.StatusOK).WithCode(InternalServerError).WithMessage(CodeMsgMap[InternalServerError]).To()
	return true
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
	return CodeMsgMap[ParamValidError]
}
