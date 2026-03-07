package gin_plus

import (
	"context"
	"sort"

	"github.com/archine/gin-plus/v4/app"
	"github.com/archine/gin-plus/v4/component/config"
)

// eventManager manages all app events with categorized storage for better performance.
// Events are sorted once and cached by type to avoid repeated type assertions.
type eventManager struct {
	// Raw events storage
	events []Event
	sorted bool

	// Cached typed events (populated after first sort)
	startingEvents              []StartingEvent
	startedEvents               []StartedEvent
	stoppedEvents               []StoppedEvent
	configEvents                []ConfigEvent
	containerRefreshBeforeEvents []ContainerRefreshBeforeEvent
	containerRefreshAfterEvents  []ContainerRefreshAfterEvent
}

func newEventManager() *eventManager {
	return &eventManager{
		events: make([]Event, 0, 8), // Pre-allocate reasonable capacity
		sorted: false,
	}
}

// register adds multiple app events to the manager
func (m *eventManager) register(events ...Event) {
	m.events = append(m.events, events...)
	m.sorted = false
	// Clear cached typed events to force re-categorization
	m.clearCache()
}

// clearCache clears all cached typed event slices
func (m *eventManager) clearCache() {
	m.startingEvents = nil
	m.startedEvents = nil
	m.stoppedEvents = nil
	m.configEvents = nil
	m.containerRefreshBeforeEvents = nil
	m.containerRefreshAfterEvents = nil
}

// ensureSorted sorts events by order and categorizes them by type for efficient access
func (m *eventManager) ensureSorted() {
	if m.sorted {
		return
	}

	if len(m.events) > 1 {
		sort.Slice(m.events, func(i, j int) bool {
			return m.events[i].Order() < m.events[j].Order()
		})
	}

	// Categorize events by type to avoid repeated type assertions
	for _, e := range m.events {
		if v, ok := e.(StartingEvent); ok {
			m.startingEvents = append(m.startingEvents, v)
		}
		if v, ok := e.(StartedEvent); ok {
			m.startedEvents = append(m.startedEvents, v)
		}
		if v, ok := e.(StoppedEvent); ok {
			m.stoppedEvents = append(m.stoppedEvents, v)
		}
		if v, ok := e.(ConfigEvent); ok {
			m.configEvents = append(m.configEvents, v)
		}
		if v, ok := e.(ContainerRefreshBeforeEvent); ok {
			m.containerRefreshBeforeEvents = append(m.containerRefreshBeforeEvents, v)
		}
		if v, ok := e.(ContainerRefreshAfterEvent); ok {
			m.containerRefreshAfterEvents = append(m.containerRefreshAfterEvents, v)
		}
	}

	m.sorted = true
}

// triggerOnStarting triggers the OnStarting event
func (m *eventManager) triggerOnStarting(ctx app.ApplicationContext) {
	m.ensureSorted()
	for _, e := range m.startingEvents {
		e.OnStarting(ctx)
	}
}

// triggerOnStarted triggers the OnStarted event
func (m *eventManager) triggerOnStarted(ctx app.ApplicationContext) {
	m.ensureSorted()
	for _, e := range m.startedEvents {
		e.OnStarted(ctx)
	}
}

// triggerOnStopped triggers the OnStopped event
func (m *eventManager) triggerOnStopped(ctx context.Context) {
	m.ensureSorted()
	for _, e := range m.stoppedEvents {
		e.OnStopped(ctx)
	}
}

// triggerConfigLoaded triggers the ConfigAfterLoad event
func (m *eventManager) triggerConfigLoaded(cp config.Provider) {
	m.ensureSorted()
	for _, e := range m.configEvents {
		e.OnConfigLoaded(cp)
	}
}

// triggerContainerRefreshBefore triggers the ContainerRefreshBefore event
func (m *eventManager) triggerContainerRefreshBefore(ctx app.ApplicationContext) {
	m.ensureSorted()
	for _, e := range m.containerRefreshBeforeEvents {
		e.OnContainerRefreshBefore(ctx)
	}
}

// triggerContainerRefreshAfter triggers the ContainerRefreshAfter event
func (m *eventManager) triggerContainerRefreshAfter(ctx app.ApplicationContext) {
	m.ensureSorted()
	for _, e := range m.containerRefreshAfterEvents {
		e.OnContainerRefreshAfter(ctx)
	}
}
