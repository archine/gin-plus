package resp

import (
	"github.com/gin-gonic/gin"
	"github.com/go-playground/validator/v10"
	log "github.com/sirupsen/logrus"
	"net/http"
	"reflect"
)

// 响应请求结果

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
	// WithMessage 设置错误信息
	WithMessage(message string) Resp
	// WithCode 设置业务状态码
	WithCode(code int) Resp
	// WithData 设置返回数据
	WithData(data interface{}) Resp
	// To 返回
	To()
}

// PaginationResult  分页结果
type PaginationResult struct {
	Total     int64       `json:"total"`      // 总条数
	PageSize  int         `json:"page_size"`  // 页大小
	PageIndex int         `json:"page_index"` // 页索引
	Data      interface{} `json:"data"`       // 数据
}

// Result 返回结果
type Result struct {
	ctx      *gin.Context `json:"-"`
	httpCode int          `json:"-"`             // http状态码
	Code     int          `json:"err_code"`      // 业务状态码,业务自己定
	Message  string       `json:"err_msg"`       // 业务提示信息
	Data     interface{}  `json:"ret,omitempty"` // 响应数据
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

// InitResp 初始化一个错误结构
func InitResp(ctx *gin.Context, httpCode int) *Result {
	return &Result{
		ctx:      ctx,
		httpCode: httpCode,
		Code:     0,
		Message:  http.StatusText(httpCode),
	}
}

// BadRequest 错误请求
// 返回true表明停止继续向下执行
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

// DataExists 数据已存在
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

// ParamInvalid 参数无效
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

// ParamValid 结构体参数校验
// 返回true表明 err 真实存在
func ParamValid(ctx *gin.Context, err error, obj interface{}) bool {
	if err == nil {
		return false
	}
	InitResp(ctx, http.StatusOK).WithCode(ParamValidError).WithMessage(getValidMsg(err, obj)).To()
	return true
}

// NoPermission 无权限
// 返回true表明 err 真实存在
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

// NoLogin 当前未登录
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

// LoginExpired 登录过期
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

// Ok 正常请求返回
func Ok(ctx *gin.Context) {
	InitResp(ctx, http.StatusOK).WithCode(0).To()
}

// Json 正常请求返回Json数据
func Json(ctx *gin.Context, data interface{}) {
	InitResp(ctx, http.StatusOK).WithCode(0).WithData(data).To()
}

// SeverError 服务器级别异常返回
func SeverError(ctx *gin.Context, err error) bool {
	if err == nil {
		return false
	}
	log.Error(err.Error())
	InitResp(ctx, http.StatusOK).WithCode(InternalServerError).WithMessage(CodeMsgMap[InternalServerError]).To()
	return true
}

// 获取校验错误信息
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
