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

const (
	NONE_LOGIN    = "Not currently logged in"
	TOKEN_EXPIRED = "Token expired"
	PARAM_FAILD   = "Parameter error"
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
		resp := InitResp(ctx, http.StatusOK).WithCode(BAD_REQUEST_CODE)
		if len(msg) > 0 {
			resp.WithMessage(msg[0])
		} else {
			resp.WithMessage(http.StatusText(http.StatusBadRequest))
		}
		resp.To()
	}
	return flag
}

// ParamInvalid parameter check type error.
// true means that flag is satisfied
func ParamInvalid(ctx *gin.Context, flag bool, msg ...string) bool {
	if flag {
		message := PARAM_FAILD
		if len(msg) > 0 {
			message = msg[0]
		}
		InitResp(ctx, http.StatusOK).WithCode(PARAM_FAILD_CODE).WithMessage(message).To()
	}
	return flag
}

// ParamValid structure parameter validation type error.
// true means that flag is satisfied
func ParamValid(ctx *gin.Context, err error, obj interface{}) bool {
	if err == nil {
		return false
	}
	InitResp(ctx, http.StatusOK).WithCode(PARAM_FAILD_CODE).WithMessage(getValidMsg(err, obj)).To()
	return true
}

// NoPermission Insufficient permission error.
// true means that flag is satisfied
func NoPermission(ctx *gin.Context, flag bool, msg ...string) bool {
	if flag {
		resp := InitResp(ctx, http.StatusOK).WithCode(NO_PERMISSION_CODE)
		if len(msg) > 0 {
			resp.WithMessage(msg[0])
		}
		resp.To()
	}
	return flag
}

// NoLogin Not logged in.
// true means that flag is satisfied
func NoLogin(ctx *gin.Context, flag bool, msg ...string) bool {
	if flag {
		message := NONE_LOGIN
		if len(msg) > 0 {
			message = msg[0]
		}
		InitResp(ctx, http.StatusUnauthorized).WithCode(NONE_LOGIN_CODE).WithMessage(message).To()
	}
	return flag
}

// LoginExpired Login expired
// true means that flag is satisfied
func LoginExpired(ctx *gin.Context, flag bool, msg ...string) bool {
	if flag {
		message := TOKEN_EXPIRED
		if len(msg) > 0 {
			message = msg[0]
		}
		InitResp(ctx, http.StatusUnauthorized).WithCode(LOGIN_TOKEN_EXPIRED).WithMessage(message).To()
	}
	return flag
}

// Ok Normal request returned, no data
func Ok(ctx *gin.Context) {
	InitResp(ctx, http.StatusOK).To()
}

// Json Normal request returned with data
func Json(ctx *gin.Context, data interface{}) {
	InitResp(ctx, http.StatusOK).WithCode(0).WithData(data).To()
}

// SeverError Server level exception
func SeverError(ctx *gin.Context, err error, msg ...string) bool {
	if err == nil {
		return false
	}
	resp := InitResp(ctx, http.StatusOK).WithCode(INTERNAL_SERVER_CODE)
	if len(msg) > 0 {
		resp.WithMessage(msg[0])
	} else {
		resp.WithMessage(http.StatusText(http.StatusInternalServerError))
	}
	resp.To()
	return true
}

// Custom User defined business code and message
func Custom(ctx *gin.Context, httpCode int, msg string) {
	InitResp(ctx, http.StatusOK).WithCode(httpCode).WithMessage(msg).To()
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
	return PARAM_FAILD
}
