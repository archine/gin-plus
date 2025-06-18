package event_manager

import (
	"github.com/archine/gin-plus/v3/event"
	"github.com/spf13/viper"
)

// AppEventManager manages all app events and triggers events at appropriate times
type AppEventManager struct {
	events []event.AppEvent
}

// NewEventManager creates a new event manager instance
func NewEventManager() *AppEventManager {
	return &AppEventManager{
		events: make([]event.AppEvent, 0),
	}
}

// Register adds multiple app events to the manager
func (em *AppEventManager) Register(events ...event.AppEvent) {
	em.events = append(em.events, events...)
}

// TriggerAppStarting triggers the AppStarting event
func (em *AppEventManager) TriggerAppStarting() {
	for _, e := range em.events {
		if appEvent, ok := e.(event.AppStartingEvent); ok {
			appEvent.OnApplicationStarting()
		}
	}
}

// TriggerAppStarted triggers the AppStarted event
func (em *AppEventManager) TriggerAppStarted() {
	for _, e := range em.events {
		if appEvent, ok := e.(event.AppStartedEvent); ok {
			appEvent.OnApplicationStarted()
		}
	}
}

// TriggerAppStopping triggers the AppStopping event
func (em *AppEventManager) TriggerAppStopping() {
	for _, e := range em.events {
		if appEvent, ok := e.(event.AppStoppingEvent); ok {
			appEvent.OnApplicationStopping()
		}
	}
}

// TriggerAppStopped triggers the AppStopped event
func (em *AppEventManager) TriggerAppStopped() {
	for _, e := range em.events {
		if appEvent, ok := e.(event.AppStoppedEvent); ok {
			appEvent.OnApplicationStopped()
		}
	}
}

// TriggerConfigBeforeLoad triggers the ConfigBeforeLoad event
func (em *AppEventManager) TriggerConfigBeforeLoad(v *viper.Viper) {
	for _, e := range em.events {
		if configEvent, ok := e.(event.ConfigBeforeLoadEvent); ok {
			configEvent.OnConfigBeforeLoad(v)
		}
	}
}

// TriggerConfigAfterLoad triggers the ConfigAfterLoad event
func (em *AppEventManager) TriggerConfigAfterLoad(v *viper.Viper) {
	for _, e := range em.events {
		if configEvent, ok := e.(event.ConfigAfterLoadEvent); ok {
			configEvent.OnConfigAfterLoad(v)
		}
	}
}

// TriggerContainerAfterInit triggers the ContainerAfterInit event
func (em *AppEventManager) TriggerContainerAfterInit() {
	for _, e := range em.events {
		if containerEvent, ok := e.(event.ContainerAfterInitEvent); ok {
			containerEvent.OnBeanContainerAfterInit()
		}
	}
}

// TriggerContainerBeforeInit triggers the ContainerBeforeInit event
func (em *AppEventManager) TriggerContainerBeforeInit() {
	for _, e := range em.events {
		if containerEvent, ok := e.(event.ContainerBeforeInitEvent); ok {
			containerEvent.OnBeanContainerBeforeInit()
		}
	}
}
