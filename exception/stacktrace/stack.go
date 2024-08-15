package stacktrace

import (
	"github.com/archine/gin-plus/v3/internal/pool"
	"runtime"
	"strconv"
	"strings"
)

var _stackPool = pool.New(func() *Stack {
	return &Stack{
		pcs: make([]uintptr, 16),
	}
})

// Stack represents a stack of program counters.
type Stack struct {
	pcs    []uintptr
	frames *runtime.Frames
}

// Capture captures a stack trace of the specified depth
func Capture(depth int) *Stack {
	stack := _stackPool.Get()
	if depth < 1 {
		stack.pcs = stack.pcs[:1]
	}
	n := runtime.Callers(3, stack.pcs)
	stack.pcs = stack.pcs[:n]
	stack.frames = runtime.CallersFrames(stack.pcs)
	return stack
}

// Next returns the next frame in the stack trace,
// and a boolean indicating whether there are more after it.
func (s *Stack) Next() (runtime.Frame, bool) {
	return s.frames.Next()
}

// Free releases resources associated with this stacktrace and back to the pool
func (s *Stack) Free() {
	s.frames = nil
	s.pcs = nil
	_stackPool.Put(s)
}

// Format formats to string
func (s *Stack) Format() string {
	var builder strings.Builder
	for {
		frame, more := s.Next()
		builder.WriteString(frame.Function)
		builder.WriteByte('\n')
		builder.WriteByte('\t')
		builder.WriteString(frame.File)
		builder.WriteByte(':')
		builder.WriteString(strconv.Itoa(frame.Line))
		if !more {
			break
		}
		builder.WriteByte('\n')
	}
	return builder.String()
}
