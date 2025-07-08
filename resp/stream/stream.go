package stream

import (
	"fmt"

	"github.com/archine/gin-plus/v4/component/pool"
	"github.com/gin-gonic/gin"
)

var _pool = pool.New(func() *Writer {
	return &Writer{}
})

// Writer represents a Server-Sent Events writer.
//
// Usage examples:
//
//	// Long connection usage (manual management)
//	writer := stream.NewWriter(ctx)
//	defer writer.Release() // Ensure to release the writer after use
//	writer.WriteEvent("xxx", "Hello")
//	writer.WriteEvent("xxx", "World")
//
//	// One-time usage (automatic management)
//	stream.Send(ctx, "message", "Hello World")
//	stream.Typewriter(ctx, "typing", "Hello World") // Character by character
//
//	// Error handling
//	if err := writer.WriteEvent("error", "Something went wrong"); err != nil {
//		log.Printf("Failed to write SSE: %v", err)
//		writer.Release()
//		return
//	}
type Writer struct {
	ctx *gin.Context
}

// NewWriter creates a new SSE writer for the given gpctx.
func NewWriter(ctx *gin.Context) *Writer {
	if ctx == nil {
		panic("gpctx cannot be nil")
	}
	w := _pool.Get()
	w.ctx = ctx
	w.ctx.Writer.Header().Set("Content-Type", "text/event-stream")
	w.ctx.Writer.Header().Set("Cache-Control", "no-cache")
	w.ctx.Writer.Header().Set("Connection", "keep-alive")
	w.ctx.Writer.Header().Set("Access-Control-Allow-Origin", "*")

	return w
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

// Release releases the writer back to the pool.
func (w *Writer) Release() {
	if w == nil {
		return
	}
	w.ctx = nil
	_pool.Put(w)
}

// Typewriter sends content character by character (typewriter effect).
func Typewriter(ctx *gin.Context, event string, content string) error {
	writer := NewWriter(ctx)
	defer writer.Release()

	return writer.WriteChars(event, content)
}

// Send sends content as a single SSE event.
func Send(ctx *gin.Context, event string, content string) error {
	writer := NewWriter(ctx)
	defer writer.Release()

	return writer.WriteEvent(event, content)
}
