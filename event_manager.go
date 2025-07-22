package gin_plus

import (
	"context"
	"github.com/archine/gin-plus/v4/app"
	"github.com/archine/gin-plus/v4/component/config"
	"sort"
)

// eventManager manages all app events with a single slice
type eventManager struct {
	events []Event
	sorted bool
}

func newEventManager() *eventManager {
	return &eventManager{
		events: make([]Event, 0),
		sorted: false,
	}
}

// register adds multiple app events to the manager
func (m *eventManager) register(events ...Event) {
	m.events = append(m.events, events...)
	m.sorted = false
}

func (m *eventManager) ensureSorted() {
	if !m.sorted && len(m.events) > 1 {
		sort.Slice(m.events, func(i, j int) bool {
			return m.events[i].Order() < m.events[j].Order()
		})
		m.sorted = true
	}
}

// triggerOnStarting triggers the OnStarting event
func (m *eventManager) triggerOnStarting() {
	m.ensureSorted()
	for _, e := range m.events {
		if lifecycleEvent, ok := e.(StartingEvent); ok {
			lifecycleEvent.OnStarting()
		}
	}
}

// triggerOnStarted triggers the OnStarted event
func (m *eventManager) triggerOnStarted() {
	m.ensureSorted()
	for _, e := range m.events {
		if lifecycleEvent, ok := e.(StartedEvent); ok {
			lifecycleEvent.OnStarted()
		}
	}
}

// triggerOnStopped triggers the OnStopped event
func (m *eventManager) triggerOnStopped(ctx context.Context) {
	m.ensureSorted()
	for _, e := range m.events {
		if lifecycleEvent, ok := e.(StoppedEvent); ok {
			lifecycleEvent.OnStopped(ctx)
		}
	}
}

// triggerConfigLoaded triggers the ConfigAfterLoad event
func (m *eventManager) triggerConfigLoaded(cp config.Provider) {
	m.ensureSorted()
	for _, e := range m.events {
		if configEvent, ok := e.(ConfigEvent); ok {
			configEvent.OnConfigLoaded(cp)
		}
	}
}

// triggerContainerRefreshBefore triggers the ContainerRefreshBefore event
func (m *eventManager) triggerContainerRefreshBefore(ctx app.ApplicationContext) {
	m.ensureSorted()
	for _, e := range m.events {
		if refreshBeforeEvent, ok := e.(ContainerRefreshBeforeEvent); ok {
			refreshBeforeEvent.OnContainerRefreshBefore(ctx)
		}
	}
}

// triggerContainerRefreshAfter triggers the ContainerRefreshAfter event
func (m *eventManager) triggerContainerRefreshAfter(ctx app.ApplicationContext) {
	m.ensureSorted()
	for _, e := range m.events {
		if refreshAfterEvent, ok := e.(ContainerRefreshAfterEvent); ok {
			refreshAfterEvent.OnContainerRefreshAfter(ctx)
		}
	}
}
