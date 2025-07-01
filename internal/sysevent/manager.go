package sysevent

import (
	"github.com/archine/gin-plus/v4/component/config"
	"github.com/archine/gin-plus/v4/component/event"
	"github.com/archine/gin-plus/v4/component/ioc"
	"sort"
)

// Manager manages all app events with a single slice
type Manager struct {
	events []event.AppEvent
	sorted bool
}

// NewEventManager creates a new event manager instance
func NewEventManager() *Manager {
	return &Manager{
		sorted: true,
	}
}

// Register adds multiple app events to the manager
func (m *Manager) Register(events ...event.AppEvent) {
	m.events = append(m.events, events...)
	m.sorted = false
}

func (m *Manager) ensureSorted() {
	if !m.sorted && len(m.events) > 1 {
		sort.Slice(m.events, func(i, j int) bool {
			return m.events[i].Order() < m.events[j].Order()
		})
		m.sorted = true
	}
}

// TriggerOnStarting triggers the OnStarting event
func (m *Manager) TriggerOnStarting() bool {
	m.ensureSorted()
	for _, e := range m.events {
		if lifecycleEvent, ok := e.(event.AppLifecycleEvent); ok {
			if !lifecycleEvent.OnStarting() {
				return false
			}
		}
	}
	return true
}

// TriggerOnStarted triggers the OnStarted event
func (m *Manager) TriggerOnStarted() {
	m.ensureSorted()
	for _, e := range m.events {
		if lifecycleEvent, ok := e.(event.AppLifecycleEvent); ok {
			lifecycleEvent.OnStarted()
		}
	}
}

// TriggerOnStopping triggers the OnStopping event
func (m *Manager) TriggerOnStopping() {
	m.ensureSorted()
	for _, e := range m.events {
		if lifecycleEvent, ok := e.(event.AppLifecycleEvent); ok {
			lifecycleEvent.OnStopping()
		}
	}
}

// TriggerOnStopped triggers the OnStopped event
func (m *Manager) TriggerOnStopped() {
	m.ensureSorted()
	for _, e := range m.events {
		if lifecycleEvent, ok := e.(event.AppLifecycleEvent); ok {
			lifecycleEvent.OnStopped()
		}
	}
}

// TriggerConfigAfterLoad triggers the ConfigAfterLoad event
func (m *Manager) TriggerConfigAfterLoad(configure config.Configure) {
	m.ensureSorted()
	for _, e := range m.events {
		if configEvent, ok := e.(event.ConfigAfterLoadEvent); ok {
			configEvent.OnConfigAfterLoad(configure)
		}
	}
}

// TriggerContainerRefreshBefore triggers the ContainerRefreshBefore event
func (m *Manager) TriggerContainerRefreshBefore(c *ioc.Container) {
	m.ensureSorted()
	for _, e := range m.events {
		if refreshBeforeEvent, ok := e.(event.ContainerRefreshBeforeEvent); ok {
			refreshBeforeEvent.OnContainerRefreshBefore(c)
		}
	}
}

// TriggerContainerRefreshAfter triggers the ContainerRefreshAfter event
func (m *Manager) TriggerContainerRefreshAfter(ct *ioc.Container) {
	m.ensureSorted()
	for _, e := range m.events {
		if refreshAfterEvent, ok := e.(event.ContainerRefreshAfterEvent); ok {
			refreshAfterEvent.OnContainerRefreshAfter(ct)
		}
	}
}
