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

type LogFormat struct {
}

func (h *LogFormat) Format(entry *log.Entry) ([]byte, error) {
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
	log.SetFormatter(&LogFormat{})
}

var printHealth = true

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
		if strings.Contains(reqUrl, "health") {
			if !printHealth {
				return
			}
			printHealth = false
		}
		businessCode, exists := c.Get("businessCode")
		if exists {
			statusCode = businessCode.(int)
		}
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
	case code == 0 || code == 200:
		return green
	default:
		return red
	}
}
