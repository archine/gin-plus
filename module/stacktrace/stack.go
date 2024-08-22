package stacktrace

import (
	"github.com/archine/gin-plus/v3/module/pool"
	"runtime"
	"strconv"
	"strings"
)

const (
	FullStack  = 16 // the size of the full stack
	CallerSkip = 3  // the number of stack frames to skip
)

var _stackPool = pool.New(func() *Stack {
	return &Stack{
		pcs: make([]uintptr, FullStack),
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
		depth = 1
	}
	stack.pcs = stack.pcs[:depth]
	n := runtime.Callers(CallerSkip, stack.pcs)
	if n < depth {
		stack.pcs = stack.pcs[:n]
	}
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

// ToString returns the string representation of the stack trace.
func (s *Stack) ToString() string {
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

// First returns the first frame in the stack trace.
func (s *Stack) First() string {
	frame, _ := s.Next()
	var sb strings.Builder
	sb.WriteString(frame.File)
	sb.WriteString(":")
	sb.WriteString(strconv.Itoa(frame.Line))
	return sb.String()
}
