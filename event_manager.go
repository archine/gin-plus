package gin_plus

import (
	"cmp"
	"context"
	"fmt"
	"reflect"
	"slices"
	"time"

	"github.com/archine/gin-plus/v4/component/config"
	"github.com/archine/gin-plus/v4/component/gplog"
)

// eventManager manages all application lifecycle events.
// It is only accessed from the goroutine running App.Run, so no locking is needed.
type eventManager struct {
	events []AppEvent
}

func newEventManager() *eventManager {
	return &eventManager{}
}

// register appends events, skipping nil values and events that are already registered,
// so repeated WithEvent calls with the same handler never fire it twice.
func (m *eventManager) register(events ...AppEvent) {
	for _, e := range events {
		if e == nil {
			continue
		}
		// Comparing interfaces panics for non-comparable dynamic types (e.g. a struct value
		// holding a map), such events are appended without deduplication.
		if reflect.TypeOf(e).Comparable() && slices.Contains(m.events, e) {
			continue
		}
		m.events = append(m.events, e)
	}
}

// trigger runs runFn for every registered event implementing T, ordered by Order().
// The sort is stable: events with equal Order keep their registration order.
func trigger[T AppEvent](m *eventManager, runFn func(T)) {
	var matched []T
	for _, c := range m.events {
		if e, ok := c.(T); ok {
			matched = append(matched, e)
		}
	}

	slices.SortStableFunc(matched, func(a, b T) int {
		return cmp.Compare(a.Order(), b.Order())
	})

	for _, handler := range matched {
		runFn(handler)
	}
}

func (m *eventManager) triggerOnStarting() {
	trigger(m, func(e StartingEvent) {
		e.OnStarting()
	})
}

func (m *eventManager) triggerOnStarted() {
	trigger(m, func(e StartedEvent) {
		e.OnStarted()
	})
}

// triggerOnStopped runs the cleanup callbacks within the given time budget.
//
// Each handler is isolated: a panic is logged and the remaining handlers still run,
// since skipping cleanup (closing pools, flushing buffers) is worse than a noisy log.
// When the budget runs out, the method returns without waiting; the callbacks observe
// the cancellation through ctx.
func (m *eventManager) triggerOnStopped(budget time.Duration) {
	ctx, cancel := context.WithTimeout(context.Background(), budget)
	defer cancel()

	done := make(chan struct{})
	go func() {
		defer close(done)
		trigger(m, func(e StoppedEvent) {
			defer func() {
				if r := recover(); r != nil {
					gplog.Error(fmt.Sprintf("OnStopped of %T panicked: %v", e, r))
				}
			}()
			e.OnStopped(ctx)
		})
	}()

	select {
	case <-done:
	case <-ctx.Done():
		gplog.Warn(fmt.Sprintf("OnStopped callbacks did not finish within exit-delay %v", budget))
	}
}

func (m *eventManager) triggerConfigLoaded(cp config.Provider) {
	trigger(m, func(e ConfigEvent) {
		e.OnConfigLoaded(cp)
	})
}

func (m *eventManager) triggerContainerRefreshBefore() {
	trigger(m, func(e ContainerRefreshBeforeEvent) {
		e.OnContainerRefreshBefore()
	})
}

func (m *eventManager) triggerContainerRefreshAfter() {
	trigger(m, func(e ContainerRefreshAfterEvent) {
		e.OnContainerRefreshAfter()
	})
}
