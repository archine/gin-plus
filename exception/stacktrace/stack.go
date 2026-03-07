package stacktrace

import (
	"runtime"
	"strconv"
	"strings"

	"github.com/archine/gin-plus/v4/component/pool"
)

const (
	MaximumStack = 32 // Maximum number of stack frames to capture and print, limiting to 32 layers
	CallerSkip   = 3  // Number of stack frames to skip, including up to the current method
)

var _stackPool = pool.New(func() *Stack {
	return &Stack{}
})

// Stack represents a stack of program counters.
type Stack struct {
	frames *runtime.Frames
}

// Capture captures a stack trace starting from the specified offset and retrieves up to the specified depth.
func Capture(skipOffset, depth int) *Stack {
	stack := _stackPool.Get()

	// Adjust depth to a valid range, defaults to FullStack if out of bounds.
	if depth < 1 || depth > MaximumStack {
		depth = MaximumStack
	}

	// Ensure skipOffset does not go below the base CallerSkip.
	if skipOffset < -CallerSkip {
		skipOffset = 0
	}

	pcs := make([]uintptr, depth)
	n := runtime.Callers(CallerSkip+skipOffset, pcs)
	if n < depth {
		pcs = pcs[:n]
	}

	stack.frames = runtime.CallersFrames(pcs)
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
	_stackPool.Put(s)
}

// ToString returns the string representation of the stack trace.
func (s *Stack) ToString() string {
	var builder strings.Builder
	// Pre-allocate reasonable capacity to reduce allocations
	builder.Grow(512)

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
	// Pre-allocate reasonable capacity
	sb.Grow(128)
	sb.WriteString(frame.File)
	sb.WriteByte(':')
	sb.WriteString(strconv.Itoa(frame.Line))
	return sb.String()
}
