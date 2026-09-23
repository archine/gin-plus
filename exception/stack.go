package exception

import (
	"runtime"
	"strconv"
	"strings"
)

// maximumStack defines the maximum number of stack frames to capture.
const maximumStack = 32

// Stack is a captured call stack. It keeps the raw program counters and resolves
// frames on demand, so it can be rendered any number of times.
type Stack struct {
	pcs []uintptr
}

// CaptureStackTrace captures up to depth frames (1-32, out-of-range values mean 32).
// skipOffset is the number of frames to skip above the caller: 0 starts the trace at the
// function calling CaptureStackTrace, 1 at its caller, and so on. Negative values mean 0.
func CaptureStackTrace(skipOffset, depth int) *Stack {
	if depth < 1 || depth > maximumStack {
		depth = maximumStack
	}
	if skipOffset < 0 {
		skipOffset = 0
	}

	pcs := make([]uintptr, depth)
	// Skip runtime.Callers and CaptureStackTrace itself.
	n := runtime.Callers(2+skipOffset, pcs)
	return &Stack{pcs: pcs[:n]}
}

// Full returns the stack trace, one "function\n\tfile:line" entry per frame.
// Leading runtime frames are skipped: when captured while recovering a panic, they belong to
// the panic machinery (runtime.gopanic, runtime.sigpanic, ...), and the trace should start
// at the line that actually failed.
func (s *Stack) Full() string {
	var builder strings.Builder
	builder.Grow(512)
	frames := runtime.CallersFrames(s.pcs)
	for more := len(s.pcs) > 0; more; {
		var frame runtime.Frame
		frame, more = frames.Next()
		if builder.Len() == 0 && more && strings.HasPrefix(frame.Function, "runtime.") {
			continue
		}
		if builder.Len() > 0 {
			builder.WriteByte('\n')
		}
		builder.WriteString(frame.Function)
		builder.WriteString("\n\t")
		builder.WriteString(frame.File)
		builder.WriteByte(':')
		builder.WriteString(strconv.Itoa(frame.Line))
	}
	return builder.String()
}

// FirstFrame returns the "file:line" of the first frame outside the runtime (see Full),
// or "" if nothing was captured.
func (s *Stack) FirstFrame() string {
	frames := runtime.CallersFrames(s.pcs)
	for more := len(s.pcs) > 0; more; {
		var frame runtime.Frame
		frame, more = frames.Next()
		if !more || !strings.HasPrefix(frame.Function, "runtime.") {
			return frame.File + ":" + strconv.Itoa(frame.Line)
		}
	}
	return ""
}
