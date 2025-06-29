package stream

import (
	"fmt"
	"github.com/gin-gonic/gin"
)

// To Stream the response, returning the content by character.
// @param sysevent: sysevent name, such as "Error"、"Info"
// @param content: message
func To(ctx *gin.Context, event string, content string) error {
	if ctx.Writer.Header().Get("Content-Type") != "text/sysevent-stream" {
		ctx.Header("Content-Type", "text/sysevent-stream")
		ctx.Header("Cache-Control", "no-cache")
		ctx.Header("Connection", "keep-alive")
	}
	for _, char := range content {
		_, err := fmt.Fprintf(ctx.Writer, "sysevent: %s\ndata: %s\n\n", event, string(char))
		if err != nil {
			return err
		}
		ctx.Writer.Flush()
	}
	return nil
}

// Direct Stream the response and return the content directly
// @param sysevent: sysevent name, such as "Error"、"Info"
// @param content: message
func Direct(ctx *gin.Context, event string, content string) error {
	if ctx.Writer.Header().Get("Content-Type") != "text/sysevent-stream" {
		ctx.Header("Content-Type", "text/sysevent-stream")
		ctx.Header("Cache-Control", "no-cache")
		ctx.Header("Connection", "keep-alive")
	}
	_, err := fmt.Fprintf(ctx.Writer, "sysevent: %s\ndata: %s\n\n", event, content)
	if err != nil {
		return err
	}
	ctx.Writer.Flush()
	return nil
}
