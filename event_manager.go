package gin_plus

import (
	"cmp"
	"context"
	"slices"

	"github.com/archine/gin-plus/v4/app"
	"github.com/archine/gin-plus/v4/component/config"
)

// eventManager manages all application lifecycle events.
type eventManager struct {
	events []AppEvent
}

func newEventManager() *eventManager {
	return &eventManager{}
}

// register adds new events to the manager. 
// It accepts a variadic list of AppEvent interfaces and appends them to the internal slice.
func (m *eventManager) register(events ...AppEvent) {
	m.events = append(m.events, events...)
}

// trigger is a generic function that triggers events of a specific type T.
// It filters the registered events to find those that match the type T, sorts them by their Order() value,
// and then executes the provided runFn callback for each matched event in the correct order.
func trigger[T AppEvent](m *eventManager, runFn func(T)) {
	var matched []T
	for _, c := range m.events {
		if e, ok := c.(T); ok {
			matched = append(matched, e)
		}
	}

	slices.SortFunc(matched, func(a, b T) int {
		return cmp.Compare(a.Order(), b.Order())
	})

	for _, handler := range matched {
		runFn(handler)
	}
}

func (m *eventManager) triggerOnStarting(ctx app.ApplicationContext) {
	trigger(m, func(e StartingEvent) {
		e.OnStarting(ctx)
	})
}

func (m *eventManager) triggerOnStarted(ctx app.ApplicationContext) {
	trigger(m, func(e StartedEvent) {
		e.OnStarted(ctx)
	})
}

func (m *eventManager) triggerOnStopped(ctx context.Context) {
	trigger(m, func(e StoppedEvent) {
		e.OnStopped(ctx)
	})
}

func (m *eventManager) triggerConfigLoaded(cp config.Provider) {
	trigger(m, func(e ConfigEvent) {
		e.OnConfigLoaded(cp)
	})
}

func (m *eventManager) triggerContainerRefreshBefore(ctx app.ApplicationContext) {
	trigger(m, func(e ContainerRefreshBeforeEvent) {
		e.OnContainerRefreshBefore(ctx)
	})
}

func (m *eventManager) triggerContainerRefreshAfter(ctx app.ApplicationContext) {
	trigger(m, func(e ContainerRefreshAfterEvent) {
		e.OnContainerRefreshAfter(ctx)
	})
}