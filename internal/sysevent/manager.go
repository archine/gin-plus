package sysevent

import (
	"sort"

	"github.com/archine/gin-plus/v4/component/ioc"
	"github.com/gin-gonic/gin"

	"github.com/archine/gin-plus/v4/component/config"
)

// Manager manages all app events with a single slice
type Manager struct {
	events []AppEvent
	sorted bool
}

// NewEventManager creates a new sysevent manager instance
func NewEventManager() *Manager {
	return &Manager{
		sorted: true,
	}
}

// Register adds multiple app events to the manager
func (m *Manager) Register(events ...AppEvent) {
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

// TriggerOnStarting triggers the OnStarting sysevent
func (m *Manager) TriggerOnStarting(engine *gin.Engine) bool {
    m.ensureSorted()
    for _, e := range m.events {
        if lifecycleEvent, ok := e.(AppLifecycleEvent); ok {
            if !lifecycleEvent.OnStarting(engine) {
                return false
            }
        }
    }
    return true
}

// TriggerOnStarted triggers the OnStarted sysevent
func (m *Manager) TriggerOnStarted(engine *gin.Engine) {
    m.ensureSorted()
    for _, e := range m.events {
        if lifecycleEvent, ok := e.(AppLifecycleEvent); ok {
            lifecycleEvent.OnStarted(engine)
        }
    }
}

// TriggerOnStopping triggers the OnStopping sysevent
func (m *Manager) TriggerOnStopping() {
    m.ensureSorted()
    for _, e := range m.events {
        if lifecycleEvent, ok := e.(AppLifecycleEvent); ok {
            lifecycleEvent.OnStopping()
        }
    }
}

// TriggerOnStopped triggers the OnStopped sysevent
func (m *Manager) TriggerOnStopped() {
    m.ensureSorted()
    for _, e := range m.events {
        if lifecycleEvent, ok := e.(AppLifecycleEvent); ok {
            lifecycleEvent.OnStopped()
        }
    }
}

// TriggerConfigAfterLoad triggers the ConfigAfterLoad sysevent
func (m *Manager) TriggerConfigAfterLoad(configure config.Configure) {
	m.ensureSorted()
	for _, e := range m.events {
		if configEvent, ok := e.(ConfigAfterLoadEvent); ok {
			configEvent.OnConfigAfterLoad(configure)
		}
	}
}

// TriggerContainerRefreshBefore triggers the ContainerRefreshBefore sysevent
func (m *Manager) TriggerContainerRefreshBefore(c *ioc.Container) {
	m.ensureSorted()
	for _, e := range m.events {
		if refreshBeforeEvent, ok := e.(ContainerRefreshBeforeEvent); ok {
			refreshBeforeEvent.OnContainerRefreshBefore(c)
		}
	}
}

// TriggerContainerRefreshAfter triggers the ContainerRefreshAfter sysevent
func (m *Manager) TriggerContainerRefreshAfter(ct *ioc.Container) {
	m.ensureSorted()
	for _, e := range m.events {
		if refreshAfterEvent, ok := e.(ContainerRefreshAfterEvent); ok {
			refreshAfterEvent.OnContainerRefreshAfter(ct)
		}
	}
}
