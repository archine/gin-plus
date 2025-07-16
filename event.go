package gin_plus

import (
	"context"
	"github.com/archine/gin-plus/v4/app"
	"github.com/archine/gin-plus/v4/component/gpconf"
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
	//   - true: Allows the application to start normally
	//   - false: Prevents the application from starting (aborts startup)
	//
	// If any listener returns false, startup is aborted immediately.
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
	//   - ctx: The context for the shutdown process, which can be used to perform graceful shutdown operations.
	OnStopped(ctx context.Context)
}

// ConfigAfterLoadEvent handles events after configuration loading completes.
// Implement this interface to perform post-loading tasks such as validation or derived value calculation.
type ConfigAfterLoadEvent interface {
	Event
	// OnConfigAfterLoad is called after all configuration files have been successfully loaded.
	OnConfigAfterLoad(cf gpconf.Configure)
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
	OnContainerRefreshBefore(ctx app.ApplicationContext)
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
	OnContainerRefreshAfter(ctx app.ApplicationContext)
}
