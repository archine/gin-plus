package gin_plus

import (
	"cmp"
	"context"
	"slices"
	"sync"

	"github.com/archine/gin-plus/v4/app"
	"github.com/archine/gin-plus/v4/component/config"
)

// Registry is a typed event bucket with lazy sorting and concurrency safety.
type Registry[T Comparable] struct {
	mu       sync.Mutex
	handlers []T
	sorted   bool
}

func (r *Registry[T]) Add(handler T) {
	r.mu.Lock()
	defer r.mu.Unlock()

	r.handlers = append(r.handlers, handler)
	r.sorted = false
}

func (r *Registry[T]) TryAdd(event any) {
	if handler, ok := event.(T); ok {
		r.Add(handler)
	}
}

func (r *Registry[T]) Trigger(runFn func(T)) {
	r.mu.Lock()
	defer r.mu.Unlock()

	if len(r.handlers) == 0 {
		return
	}

	if !r.sorted {
		slices.SortFunc(r.handlers, func(a, b T) int {
			return cmp.Compare(a.Order(), b.Order())
		})
		r.sorted = true
	}

	for _, handler := range r.handlers {
		runFn(handler)
	}
}

type eventManager struct {
	starting      Registry[StartingEvent]
	started       Registry[StartedEvent]
	stopped       Registry[StoppedEvent]
	config        Registry[ConfigEvent]
	refreshBefore Registry[ContainerRefreshBeforeEvent]
	refreshAfter  Registry[ContainerRefreshAfterEvent]

	registries []interface{ TryAdd(any) }
}

func newEventManager() *eventManager {
	m := &eventManager{}
	m.registries = []interface{ TryAdd(any) }{
		&m.starting,
		&m.started,
		&m.stopped,
		&m.config,
		&m.refreshBefore,
		&m.refreshAfter,
	}
	return m
}

func (m *eventManager) register(events ...Comparable) {
	for _, event := range events {
		for _, registry := range m.registries {
			registry.TryAdd(event)
		}
	}
}

func (m *eventManager) triggerOnStarting(ctx app.ApplicationContext) {
	m.starting.Trigger(func(e StartingEvent) {
		e.OnStarting(ctx)
	})
}

func (m *eventManager) triggerOnStarted(ctx app.ApplicationContext) {
	m.started.Trigger(func(e StartedEvent) {
		e.OnStarted(ctx)
	})
}

func (m *eventManager) triggerOnStopped(ctx context.Context) {
	m.stopped.Trigger(func(e StoppedEvent) {
		e.OnStopped(ctx)
	})
}

func (m *eventManager) triggerConfigLoaded(cp config.Provider) {
	m.config.Trigger(func(e ConfigEvent) {
		e.OnConfigLoaded(cp)
	})
}

func (m *eventManager) triggerContainerRefreshBefore(ctx app.ApplicationContext) {
	m.refreshBefore.Trigger(func(e ContainerRefreshBeforeEvent) {
		e.OnContainerRefreshBefore(ctx)
	})
}

func (m *eventManager) triggerContainerRefreshAfter(ctx app.ApplicationContext) {
	m.refreshAfter.Trigger(func(e ContainerRefreshAfterEvent) {
		e.OnContainerRefreshAfter(ctx)
	})
}
