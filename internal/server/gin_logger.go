package server

import "strings"

// logWriter adapts a gplog level function to io.Writer, e.g. logWriter(gplog.Info).
// It bridges writers such as http.Server.ErrorLog into gplog; request logging is handled by
// middleware.Logger, which needs the *gin.Context and therefore writes through gplog directly.
type logWriter func(text string, keyvals ...any)

func (w logWriter) Write(p []byte) (int, error) {
	// Always report the full length: a short write would make callers treat it as an error.
	if s := strings.TrimSpace(string(p)); s != "" {
		w(s)
	}
	return len(p), nil
}
