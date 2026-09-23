package stream

import (
	"io"
	"strings"

	"github.com/gin-gonic/gin"
)

var (
	// dataReplacer splits multi-line data into consecutive "data:" lines. CR, LF and CRLF are
	// all line terminators in SSE, so any of them would otherwise end the field early.
	dataReplacer = strings.NewReplacer("\r\n", "\ndata: ", "\r", "\ndata: ", "\n", "\ndata: ")

	// eventReplacer drops line terminators from event names, which cannot span lines.
	eventReplacer = strings.NewReplacer("\r", "", "\n", "")
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

	h := ctx.Writer.Header()
	h.Set("Content-Type", "text/event-stream")
	h.Set("Cache-Control", "no-cache")
	h.Set("Connection", "keep-alive")
	// Disable proxy buffering (nginx), otherwise events are delivered in batches.
	h.Set("X-Accel-Buffering", "no")

	return &Writer{ctx: ctx}
}

// WriteEvent writes a single SSE event. Line breaks in data are preserved.
func (w *Writer) WriteEvent(event, data string) error {
	var b strings.Builder
	b.Grow(len(event) + len(data) + 16)
	if event != "" {
		b.WriteString("event: ")
		eventReplacer.WriteString(&b, event)
		b.WriteByte('\n')
	}
	b.WriteString("data: ")
	dataReplacer.WriteString(&b, data)
	b.WriteString("\n\n")

	// One write per event, so a failure never leaves a half-written frame behind.
	if _, err := io.WriteString(w.ctx.Writer, b.String()); err != nil {
		return err
	}
	w.ctx.Writer.Flush()
	return nil
}

// WriteData writes data without an event name.
func (w *Writer) WriteData(data string) error {
	return w.WriteEvent("", data)
}

// WriteChars writes content character by character (for typewriter effect).
// It stops with the context error as soon as the client disconnects.
func (w *Writer) WriteChars(event, content string) error {
	reqCtx := w.ctx.Request.Context()
	for _, char := range content {
		if err := reqCtx.Err(); err != nil {
			return err
		}
		if err := w.WriteEvent(event, string(char)); err != nil {
			return err
		}
	}
	return nil
}

// Typewriter sends content character by character (typewriter effect).
func Typewriter(ctx *gin.Context, event string, content string) error {
	return NewWriter(ctx).WriteChars(event, content)
}

// Send sends content as a single SSE event.
func Send(ctx *gin.Context, event string, content string) error {
	return NewWriter(ctx).WriteEvent(event, content)
}
