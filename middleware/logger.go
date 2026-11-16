package middleware

import (
	"fmt"
	"strings"
	"time"

	"github.com/archine/gin-plus/v4/component/gplog"
	"github.com/gin-gonic/gin"
)

// requestLog holds the fields gathered from a request that has finished its handler chain.
type requestLog struct {
	status   int
	latency  time.Duration
	clientIP string
	method   string
	path     string
	errMsg   string
}

// Logger returns a middleware that writes one entry per request through gplog.
//
// It replaces Gin's LoggerWithConfig so that request logs and the logs written inside
// handlers share a single source of truth: the logger's configured keys (gin-plus.log.keys)
// are appended by gplog itself, reading them from the *gin.Context. Because the engine runs
// with ContextWithFallback, those keys are looked up in the Gin context (c.Set) first and in
// the request context (c.Request.WithContext) second, so either style works.
//
// Level follows the response status: 5xx is logged as error, 4xx as warn, everything else as info.
// skipPaths excludes matching requests from the log, see skipMatcher for the accepted forms.
func Logger(skipPaths []string) gin.HandlerFunc {
	skip := newSkipMatcher(skipPaths)
	// The log format cannot change while the application runs, so resolve it once here
	// rather than on every request.
	jsonFormat := gplog.GetLogger().GetFormat() == gplog.JSONFormat

	return func(ctx *gin.Context) {
		start := time.Now()
		rawQuery := ctx.Request.URL.RawQuery

		ctx.Next()

		// Skip rules are evaluated after the handlers run: it costs nothing there and keeps
		// the matcher a pure read of the finished request.
		if skip.match(ctx) {
			return
		}

		rl := requestLog{
			status:   ctx.Writer.Status(),
			latency:  time.Since(start),
			clientIP: ctx.ClientIP(),
			method:   ctx.Request.Method,
			path:     ctx.Request.URL.Path,
			errMsg:   ctx.Errors.ByType(gin.ErrorTypePrivate).String(),
		}
		if rawQuery != "" {
			rl.path += "?" + rawQuery
		}

		writeRequestLog(ctx, &rl, jsonFormat)
	}
}

// writeRequestLog renders rl and hands it to gplog at the level matching the status code.
// The context is passed along so the logger can append its configured keys.
func writeRequestLog(ctx *gin.Context, rl *requestLog, jsonFormat bool) {
	var text string
	var keyvals []any

	if jsonFormat {
		text = "HTTP Request"
		keyvals = []any{
			"status", rl.status,
			"latency", rl.latency.String(),
			"client_ip", rl.clientIP,
			"req_path", rl.path,
			"req_method", rl.method,
		}
	} else {
		text = consoleLine(rl)
	}
	// ErrorMsgs.String separates entries with a newline; collapse it so a console line stays
	// one line and every format keeps the same field value.
	if rl.errMsg != "" {
		keyvals = append(keyvals, "error", strings.Join(strings.Fields(rl.errMsg), " "))
	}

	switch {
	case rl.status >= 500:
		gplog.ErrorWithCtx(ctx, text, keyvals...)
	case rl.status >= 400:
		gplog.WarnWithCtx(ctx, text, keyvals...)
	default:
		gplog.InfoWithCtx(ctx, text, keyvals...)
	}
}

// consoleLine renders the compact, column aligned line used by the console format.
// No ANSI color is applied: the same text is written to every gplog core, including the
// rotating file, where escape sequences would corrupt the output. The level, which gplog
// colors on the console only, still carries the visual signal.
func consoleLine(rl *requestLog) string {
	latency := rl.latency
	if latency > time.Minute {
		latency = latency.Truncate(time.Second)
	}
	return fmt.Sprintf("| %3d | %13v | %15s | %-7s %s",
		rl.status, latency, rl.clientIP, rl.method, rl.path)
}

// skipMatcher decides whether a request is excluded from request logging.
//
// Entries are compared verbatim against c.FullPath(), the route template Gin matched, so
// "/user/:id" skips every value of that parameter and "/health" skips that exact route.
// No wildcard or prefix form exists: the set of templates is small and known up front, and a
// plain set lookup keeps matching O(1) without a pattern engine.
//
// A request that matched no route has an empty FullPath, so 404s are never skipped.
type skipMatcher struct {
	paths map[string]struct{}
}

// newSkipMatcher compiles the configured paths once, so matching costs no allocation per request.
func newSkipMatcher(paths []string) *skipMatcher {
	if len(paths) == 0 {
		return nil
	}
	m := &skipMatcher{paths: make(map[string]struct{}, len(paths))}
	for _, path := range paths {
		m.paths[path] = struct{}{}
	}
	return m
}

// match reports whether the request should be skipped.
func (m *skipMatcher) match(ctx *gin.Context) bool {
	if m == nil {
		return false
	}
	full := ctx.FullPath()
	if full == "" {
		return false
	}
	_, ok := m.paths[full]
	return ok
}
