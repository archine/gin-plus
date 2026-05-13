package gin_plus

import "sync"

// ShutdownSignal exposes an application-level shutdown notification.
// Long-lived handlers and workers can listen on Done to exit before the
// HTTP server waits for all active requests to complete.
type ShutdownSignal interface {
	// Done returns a receive-only channel that is closed when the application
	// starts shutting down.
	Done() <-chan struct{}
}

type shutdownSignal struct {
	done chan struct{}
	once sync.Once
}

func newShutdownSignal() *shutdownSignal {
	return &shutdownSignal{
		done: make(chan struct{}),
	}
}

func (s *shutdownSignal) Done() <-chan struct{} {
	return s.done
}

func (s *shutdownSignal) close() {
	s.once.Do(func() {
		close(s.done)
	})
}
