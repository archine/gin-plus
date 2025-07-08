package gin_plus

import (
	"context"
	"github.com/archine/gin-plus/v4/app"
	"github.com/archine/gin-plus/v4/component/config"
	"sort"
)

// Event is the base interface for all app events.
type Event interface {
	// Order returns the order of the event.
	// Lower numbers indicate higher priority.
	Order() int
}

// LifecycleEvent is the base interface for all application lifecycle events.
// Implement this interface to hook into the application's start and stop lifecycle phases.
type LifecycleEvent interface {
	Event

	// OnStarting is called before the application starts.
	// Use this method to perform initialization tasks, such as:
	//   - Validating configuration
	//   - Initializing external connections (databases, caches, etc.)
	//   - Setting up monitoring and health checks
	//   - Performing pre-flight checks
	//   - Registering middleware or routes
	// Return value:
	//   - true: Prevents the application from starting (aborts startup)
	//   - false: Allows the application to start normally
	OnStarting() bool

	// OnStarted is called after the application has started successfully.
	// Use this method to perform post-start tasks, such as:
	//   - Starting background workers or schedulers
	//   - Sending startup notifications
	//   - Performing warm-up operations
	//   - Logging startup completion status
	//   - Triggering external system notifications
	// Note: At this point, the HTTP server is running and ready to accept requests.
	OnStarted()

	// OnStopped is called after the HTTP server has stopped accepting new requests.
	// Use this method to perform cleanup tasks, such as:
	//   - Closing database connections
	//   - Stopping background workers or schedulers
	//   - Releasing resources
	// Args:
	//   - ctx: The gpctx for the shutdown process, which can be used to perform graceful shutdown operations.
	OnStopped(ctx context.Context)
}

// ConfigAfterLoadEvent handles events after configuration loading completes.
// Implement this interface to perform post-loading tasks such as validation or derived value calculation.
type ConfigAfterLoadEvent interface {
	Event
	// OnConfigAfterLoad is called after all configuration files have been successfully loaded.
	OnConfigAfterLoad(cf config.Configure)
}

// ContainerRefreshBeforeEvent handles events triggered before container refresh.
// Implement this interface to perform pre-refresh setup and configuration tasks.
type ContainerRefreshBeforeEvent interface {
	Event
	// OnContainerRefreshBefore is called before the IoC container begins its refresh process.
	// This is the ideal place to perform container preparation tasks such as:
	//
	//  - Registering additional bean definitions
	//  - Registering pre-configured bean instances
	//  - Modifying container configuration
	//  - Performing pre-refresh validations
	//  - Setting up custom bean processors
	OnContainerRefreshBefore(ctx *app.Context)
}

// ContainerRefreshAfterEvent handles events triggered after container refresh completion.
// Implement this interface to perform post-refresh finalization and validation tasks.
type ContainerRefreshAfterEvent interface {
	Event
	// OnContainerRefreshAfter is called after the IoC container has completed its refresh process.
	// At this point, all beans have been created, dependencies injected, and the container is ready for use.
	// This is the ideal place to perform post-refresh tasks such as:
	//
	//  - Validating bean initialization results
	//  - Performing cross-bean validations
	//  - Starting background services that depend on beans
	//  - Logging container status and statistics
	//  - Triggering application-specific initialization logic
	//
	// Note: The container state should not be modified at this point.
	OnContainerRefreshAfter(ctx *app.Context)
}

// eventManager manages all app events with a single slice
type eventManager struct {
	events []Event
	sorted bool
}

// Register adds multiple app events to the manager
func (m *eventManager) Register(events ...Event) {
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

// TriggerOnStarting triggers the OnStarting event
func (m *eventManager) TriggerOnStarting() bool {
	m.ensureSorted()
	for _, e := range m.events {
		if lifecycleEvent, ok := e.(LifecycleEvent); ok {
			if !lifecycleEvent.OnStarting() {
				return false
			}
		}
	}
	return true
}

// TriggerOnStarted triggers the OnStarted event
func (m *eventManager) TriggerOnStarted() {
	m.ensureSorted()
	for _, e := range m.events {
		if lifecycleEvent, ok := e.(LifecycleEvent); ok {
			lifecycleEvent.OnStarted()
		}
	}
}

// TriggerOnStopped triggers the OnStopped event
func (m *eventManager) TriggerOnStopped(ctx context.Context) {
	m.ensureSorted()
	for _, e := range m.events {
		if lifecycleEvent, ok := e.(LifecycleEvent); ok {
			lifecycleEvent.OnStopped(ctx)
		}
	}
}

// TriggerConfigAfterLoad triggers the ConfigAfterLoad event
func (m *eventManager) TriggerConfigAfterLoad(configure config.Configure) {
	m.ensureSorted()
	for _, e := range m.events {
		if configEvent, ok := e.(ConfigAfterLoadEvent); ok {
			configEvent.OnConfigAfterLoad(configure)
		}
	}
}

// TriggerContainerRefreshBefore triggers the ContainerRefreshBefore event
func (m *eventManager) TriggerContainerRefreshBefore(ctx *app.Context) {
	m.ensureSorted()
	for _, e := range m.events {
		if refreshBeforeEvent, ok := e.(ContainerRefreshBeforeEvent); ok {
			refreshBeforeEvent.OnContainerRefreshBefore(ctx)
		}
	}
}

// TriggerContainerRefreshAfter triggers the ContainerRefreshAfter event
func (m *eventManager) TriggerContainerRefreshAfter(ctx *app.Context) {
	m.ensureSorted()
	for _, e := range m.events {
		if refreshAfterEvent, ok := e.(ContainerRefreshAfterEvent); ok {
			refreshAfterEvent.OnContainerRefreshAfter(ctx)
		}
	}
}
