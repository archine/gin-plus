package gin_plus

import (
	"context"

	"github.com/archine/gin-plus/v4/app"
	"github.com/archine/gin-plus/v4/component/config"
)

// Comparable is the base ordering contract for all app events.
type Comparable interface {
	// Order returns the order of the event.
	// Lower numbers indicate higher priority.
	Order() int
}

// Event is kept as a compatibility alias.
// Deprecated: use Comparable.
type Event = Comparable

// StartingEvent is called before the application starts.
// Implement this interface to perform initialization tasks.
type StartingEvent interface {
	Comparable

	// OnStarting is called before the application starts.
	// Use this method to perform initialization tasks, such as:
	//   - Validating configuration
	//   - Initializing external connections (databases, caches, etc.)
	//   - Setting up monitoring and health checks
	//   - Performing pre-flight checks
	//   - Registering middleware or routes
	OnStarting(ctx app.ApplicationContext)
}

// StartedEvent is called after the application has started successfully.
// Implement this interface to perform post-start tasks.
type StartedEvent interface {
	Comparable

	// OnStarted is called after the application has started successfully.
	// Use this method to perform post-start tasks, such as:
	//   - Starting background workers or schedulers
	//   - Sending startup notifications
	//   - Performing warm-up operations
	//   - Logging startup completion status
	//   - Triggering external system notifications
	// Note: At this point, the HTTP server is running and ready to accept requests.
	OnStarted(ctx app.ApplicationContext)
}

// StoppedEvent is called after the HTTP server has stopped.
// Implement this interface to perform cleanup tasks.
type StoppedEvent interface {
	Comparable

	// OnStopped is called after the HTTP server has stopped accepting new requests.
	// Use this method to perform cleanup tasks, such as:
	//   - Closing database connections
	//   - Stopping background workers or schedulers
	//   - Releasing resources
	// Args:
	//   - ctx: The context for the shutdown process, which can be used to perform graceful shutdown operations.
	OnStopped(ctx context.Context)
}

// ConfigEvent is the interface for events related to configuration loading.
type ConfigEvent interface {
	Comparable
	// OnConfigLoaded is called after all configuration files have been successfully loaded.
	OnConfigLoaded(cp config.Provider)
}

// ContainerRefreshBeforeEvent is called before the IoC container refresh process.
// Implement this interface to perform container preparation tasks.
type ContainerRefreshBeforeEvent interface {
	Comparable
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

// ContainerRefreshAfterEvent is called after the IoC container refresh process.
// Implement this interface to perform post-refresh tasks.
type ContainerRefreshAfterEvent interface {
	Comparable
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
