package event

import (
	"github.com/spf13/viper"
)

// Manager manages all app events and triggers events at appropriate times
type Manager struct {
	events []AppEvent
}

// NewEventManager creates a new event manager instance
func NewEventManager() *Manager {
	return &Manager{
		events: make([]AppEvent, 0),
	}
}

// Register adds multiple app events to the manager
func (m *Manager) Register(events ...AppEvent) {
	m.events = append(m.events, events...)
}

// TriggerAppStarting triggers the AppStarting event
func (m *Manager) TriggerAppStarting() {
	for _, e := range m.events {
		if appEvent, ok := e.(AppStartingEvent); ok {
			appEvent.OnApplicationStarting()
		}
	}
}

// TriggerAppStarted triggers the AppStarted event
func (m *Manager) TriggerAppStarted() {
	for _, e := range m.events {
		if appEvent, ok := e.(AppStartedEvent); ok {
			appEvent.OnApplicationStarted()
		}
	}
}

// TriggerAppStopping triggers the AppStopping event
func (m *Manager) TriggerAppStopping() {
	for _, e := range m.events {
		if appEvent, ok := e.(AppStoppingEvent); ok {
			appEvent.OnApplicationStopping()
		}
	}
}

// TriggerAppStopped triggers the AppStopped event
func (m *Manager) TriggerAppStopped() {
	for _, e := range m.events {
		if appEvent, ok := e.(AppStoppedEvent); ok {
			appEvent.OnApplicationStopped()
		}
	}
}

// TriggerConfigBeforeLoad triggers the ConfigBeforeLoad event
func (m *Manager) TriggerConfigBeforeLoad(v *viper.Viper) {
	for _, e := range m.events {
		if configEvent, ok := e.(ConfigBeforeLoadEvent); ok {
			configEvent.OnConfigBeforeLoad(v)
		}
	}
}

// TriggerConfigAfterLoad triggers the ConfigAfterLoad event
func (m *Manager) TriggerConfigAfterLoad(v *viper.Viper) {
	for _, e := range m.events {
		if configEvent, ok := e.(ConfigAfterLoadEvent); ok {
			configEvent.OnConfigAfterLoad(v)
		}
	}
}
