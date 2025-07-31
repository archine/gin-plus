package server

import (
	"fmt"
	"github.com/archine/gin-plus/v4/component/gplog"
	"github.com/gin-gonic/gin"
	"strings"
	"time"
)

// logWriter implements gin.LoggerOutput interface
type logWriter struct{}

func (l *logWriter) Write(p []byte) (n int, err error) {
	s := strings.TrimSpace(string(p))
	if s == "" {
		return 0, nil
	}

	gplog.Info(s)
	return len(p), nil
}

// consoleFormatter is the default log format function Logger middleware uses.
var consoleFormatter = func(param gin.LogFormatterParams) string {
	var statusColor, methodColor, resetColor string
	if param.IsOutputColor() {
		statusColor = param.StatusCodeColor()
		methodColor = param.MethodColor()
		resetColor = param.ResetColor()
	}

	if param.Latency > time.Minute {
		param.Latency = param.Latency.Truncate(time.Second)
	}
	return fmt.Sprintf("|%s %3d %s| %13v | %15s |%s %-7s %s %s\n%s",
		statusColor, param.StatusCode, resetColor,
		param.Latency,
		param.ClientIP,
		methodColor, param.Method, resetColor,
		param.Path,
		param.ErrorMessage,
	)
}

// jsonFormatter is the default log format function Logger middleware uses for JSON output.
var jsonFormatter = func(param gin.LogFormatterParams) string {
	if param.Latency > time.Minute {
		param.Latency = param.Latency.Truncate(time.Second)
	}

	fields := make([]gplog.Field, 0, 5+len(param.Keys))
	fields = append(fields,
		gplog.Field{Key: "status", Value: param.StatusCode},
		gplog.Field{Key: "latency", Value: param.Latency.String()},
		gplog.Field{Key: "client_ip", Value: param.ClientIP},
		gplog.Field{Key: "req_path", Value: param.Path},
		gplog.Field{Key: "req_method", Value: param.Method},
	)

	if param.Keys != nil {
		for k, v := range param.Keys {
			fields = append(fields, gplog.Field{Key: k, Value: v})
		}
	}

	switch {
	case param.StatusCode >= 500:
		gplog.Error("HTTP Request", fields...)
	case param.StatusCode >= 400:
		gplog.Warn("HTTP Request", fields...)
	default:
		gplog.Info("HTTP Request", fields...)
	}

	return ""
}
