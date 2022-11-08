package plugin

import (
	"fmt"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"net/http"
	"strings"
	"time"
)

// Logger middle ware

// font color
const (
	red     = "\033[31m"
	cyan    = "\033[36m"
	yellow  = "\033[33m"
	blue    = "\033[34m"
	white   = "\033[37m"
	green   = "\033[32m"
	magenta = "\033[35m"
	reset   = "\033[0m"
)

type HjLogFormat struct {
}

func (h *HjLogFormat) Format(entry *log.Entry) ([]byte, error) {
	timestamp := entry.Time.Format("2006-01-02 15:04:05")
	sprintf := fmt.Sprintf("====> [%s] %s[%s]%s %s \n", timestamp, getLevelColor(entry.Level), strings.ToUpper(entry.Level.String()), reset, entry.Message)
	return []byte(sprintf), nil
}

// InitLog Initialize logger level and time format
func InitLog(levelString string) {
	level, err := log.ParseLevel(levelString)
	if err != nil {
		log.Fatalf("init logger module failed, invalid logger level: %s", levelString)
	}
	log.SetLevel(level)
	log.SetFormatter(&HjLogFormat{})
}

// LogMiddleware Log print
func LogMiddleware() gin.HandlerFunc {
	return func(c *gin.Context) {
		startTime := time.Now()
		c.Next()
		endTime := time.Now()
		processTime := endTime.Sub(startTime)
		reqMethod := c.Request.Method
		reqUrl := c.Request.RequestURI
		statusCode := c.Writer.Status()
		log.Infof("| %s %3d %s | %13v |%s %s %s  %s ", getStatusColor(statusCode), statusCode, reset, processTime, getMethodColor(reqMethod), reqMethod, reset, reqUrl)
	}
}

func getLevelColor(level log.Level) string {
	switch level {
	case log.DebugLevel, log.TraceLevel:
		return cyan
	case log.ErrorLevel, log.FatalLevel:
		return red
	case log.WarnLevel:
		return yellow
	default:
		return green
	}
}

func getMethodColor(method string) string {
	switch method {
	case http.MethodGet:
		return blue
	case http.MethodPost:
		return cyan
	case http.MethodPut:
		return yellow
	case http.MethodDelete:
		return red
	case http.MethodPatch:
		return green
	case http.MethodHead:
		return magenta
	case http.MethodOptions:
		return white
	default:
		return reset
	}
}

func getStatusColor(code int) string {
	switch {
	case code >= http.StatusOK && code < http.StatusMultipleChoices:
		return green
	case code >= http.StatusMultipleChoices && code < http.StatusBadRequest:
		return white
	case code >= http.StatusBadRequest && code < http.StatusInternalServerError:
		return yellow
	default:
		return red
	}
}
