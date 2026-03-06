package stream

import (
	"fmt"

	"github.com/gin-gonic/gin"
)

// Writer represents a Server-Sent Events writer.
//
// Usage examples:
//
//	// Long connection usage
//	writer := stream.NewWriter(ctx)
//	writer.WriteEvent("message", "Hello")
//	writer.WriteEvent("message", "World")
//
//	// One-time usage
//	stream.Send(ctx, "message", "Hello World")
//	stream.Typewriter(ctx, "typing", "Hello World") // Character by character
//
//	// Error handling
//	if err := writer.WriteEvent("error", "Something went wrong"); err != nil {
//		log.Printf("Failed to write SSE: %v", err)
//		return
//	}
type Writer struct {
	ctx *gin.Context
}

// NewWriter creates a new SSE writer and sets up the necessary headers.
func NewWriter(ctx *gin.Context) *Writer {
	if ctx == nil {
		panic("ctx cannot be nil")
	}

	ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	ctx.Writer.Header().Set("Cache-Control", "no-cache")
	ctx.Writer.Header().Set("Connection", "keep-alive")
	ctx.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	return &Writer{ctx: ctx}
}

// WriteEvent writes a single SSE event.
func (w *Writer) WriteEvent(event, data string) error {
	_, err := fmt.Fprintf(w.ctx.Writer, "event: %s\ndata: %s\n\n", event, data)
	if err != nil {
		return err
	}
	w.ctx.Writer.Flush()
	return nil
}

// WriteData writes data without an event name.
func (w *Writer) WriteData(data string) error {
	_, err := fmt.Fprintf(w.ctx.Writer, "data: %s\n\n", data)
	if err != nil {
		return err
	}
	w.ctx.Writer.Flush()
	return nil
}

// WriteChars writes content character by character (for typewriter effect).
func (w *Writer) WriteChars(event, content string) error {
	for _, char := range content {
		if err := w.WriteEvent(event, string(char)); err != nil {
			return err
		}
	}
	return nil
}

// Typewriter sends content character by character (typewriter effect).
func Typewriter(ctx *gin.Context, event string, content string) error {
	writer := NewWriter(ctx)
	return writer.WriteChars(event, content)
}

// Send sends content as a single SSE event.
func Send(ctx *gin.Context, event string, content string) error {
	writer := NewWriter(ctx)
	return writer.WriteEvent(event, content)
}
