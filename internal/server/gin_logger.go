package server

import (
	"fmt"
	"strings"
	"time"

	"github.com/archine/gin-plus/v4/component/gplog"
	"github.com/gin-gonic/gin"
)

// logWriter adapts a gplog level function to io.Writer, e.g. logWriter(gplog.Info).
// It bridges writers such as the Gin logger output and http.Server.ErrorLog into gplog.
type logWriter func(text string, keyvals ...any)

func (w logWriter) Write(p []byte) (int, error) {
	// Always report the full length: a short write would make callers treat it as an error.
	if s := strings.TrimSpace(string(p)); s != "" {
		w(s)
	}
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

	keyvals := []any{
		"status", param.StatusCode,
		"latency", param.Latency.String(),
		"client_ip", param.ClientIP,
		"req_path", param.Path,
		"req_method", param.Method,
	}

	// Keys is map[any]any since Gin 1.12; non-string keys are stringified instead of asserted.
	for k, v := range param.Keys {
		key, ok := k.(string)
		if !ok {
			key = fmt.Sprint(k)
		}
		keyvals = append(keyvals, key, v)
	}

	switch {
	case param.StatusCode >= 500:
		gplog.Error("HTTP Request", keyvals...)
	case param.StatusCode >= 400:
		gplog.Warn("HTTP Request", keyvals...)
	default:
		gplog.Info("HTTP Request", keyvals...)
	}

	return ""
}
