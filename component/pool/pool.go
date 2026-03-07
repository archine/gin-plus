package pool

import "sync"

// Pool is a type-safe wrapper around sync.Pool that provides generic object pooling.
// It eliminates the need for type assertions and provides a cleaner API for reusing objects.
//
// Example:
//
//	type Buffer struct {
//	    data []byte
//	}
//
//	bufferPool := pool.New(func() *Buffer {
//	    return &Buffer{data: make([]byte, 0, 1024)}
//	})
//
//	buf := bufferPool.Get()
//	defer bufferPool.Put(buf)
//	// use buf...
type Pool[T any] struct {
	pool sync.Pool
}

// New creates a new type-safe Pool for type T.
// The provided function fn will be called to construct new instances when the pool is empty.
//
// Parameters:
//   - fn: Factory function that creates new instances of T
//
// Returns: A new Pool instance
//
// Example:
//
//	stringPool := pool.New(func() *string {
//	    s := ""
//	    return &s
//	})
func New[T any](fn func() T) *Pool[T] {
	return &Pool[T]{
		pool: sync.Pool{
			New: func() any {
				return fn()
			},
		},
	}
}

// Get retrieves an object from the pool or creates a new one if the pool is empty.
// The returned object should be returned to the pool via Put() when no longer needed.
//
// Returns: An instance of T from the pool or newly created
//
// Example:
//
//	obj := myPool.Get()
//	defer myPool.Put(obj)
func (p *Pool[T]) Get() T {
	return p.pool.Get().(T)
}

// Put returns an object to the pool for reuse.
// The object should not be used after being returned to the pool.
//
// Parameters:
//   - x: The object to return to the pool
//
// Note: It's the caller's responsibility to reset the object's state before returning it.
//
// Example:
//
//	buf := bufferPool.Get()
//	// use buf...
//	buf.Reset() // Reset state before returning
//	bufferPool.Put(buf)
func (p *Pool[T]) Put(x T) {
	p.pool.Put(x)
}

// GetWithReset retrieves an object from the pool and applies a reset function to it.
// This is useful when you want to ensure objects are always in a clean state.
//
// Parameters:
//   - reset: Function to reset the object's state
//
// Returns: A reset instance of T
//
// Example:
//
//	buf := bufferPool.GetWithReset(func(b *Buffer) {
//	    b.data = b.data[:0]
//	})
func (p *Pool[T]) GetWithReset(reset func(T)) T {
	obj := p.pool.Get().(T)
	if reset != nil {
		reset(obj)
	}
	return obj
}

// PutWithReset returns an object to the pool after applying a reset function.
// This ensures the object is in a clean state before being reused.
//
// Parameters:
//   - x: The object to return to the pool
//   - reset: Function to reset the object's state
//
// Example:
//
//	bufferPool.PutWithReset(buf, func(b *Buffer) {
//	    b.data = b.data[:0]
//	})
func (p *Pool[T]) PutWithReset(x T, reset func(T)) {
	if reset != nil {
		reset(x)
	}
	p.pool.Put(x)
}
