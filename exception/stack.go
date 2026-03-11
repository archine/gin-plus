package exception

import (
    "runtime"
    "strconv"
    "strings"
)

// MaximumStack defines the maximum number of stack frames to capture.
const maximumStack = 32

// CallerSkip defines the number of stack frames to skip, including the current method.
const callerSkip = 3

// Stack represents a stack of program counters.
type Stack struct {
    frames *runtime.Frames
}

// CaptureStackTrace captures a stack trace starting from the specified offset and retrieves up to the specified depth.
func CaptureStackTrace(skipOffset, depth int) *Stack {
    // Adjust depth to a valid range, defaults to maximumStack if out of bounds.
    if depth < 1 || depth > maximumStack {
        depth = maximumStack
    }

    // Ensure skipOffset does not go below the base callerSkip.
    if skipOffset < -callerSkip {
        skipOffset = 0
    }

    pcs := make([]uintptr, depth)
    n := runtime.Callers(callerSkip+skipOffset, pcs)
    if n < depth {
        pcs = pcs[:n]
    }

    return &Stack{frames: runtime.CallersFrames(pcs)}
}

// Next returns the next frame in the stack trace,
// and a boolean indicating whether there are more after it.
func (s *Stack) Next() (runtime.Frame, bool) {
    return s.frames.Next()
}

// Full returns the complete string representation of the stack trace
func (s *Stack) Full() string {
    var builder strings.Builder
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

// FirstFrame returns the first frame in the stack trace.
func (s *Stack) FirstFrame() string {
    frame, _ := s.Next()
    var sb strings.Builder
    sb.Grow(128)
    sb.WriteString(frame.File)
    sb.WriteByte(':')
    sb.WriteString(strconv.Itoa(frame.Line))
    return sb.String()
}