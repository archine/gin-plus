package event

import (
	"github.com/archine/gin-plus/v4/component/config"
	"github.com/archine/gin-plus/v4/component/ioc"
)

// AppEvent is the base interface for all app events.
type AppEvent interface {
	// Order returns the order of the sysevent.
	// Lower numbers indicate higher priority.
	Order() int
}

// AppLifecycleEvent is the base interface for all application lifecycle events.
// Implement this interface to hook into the application's start and stop lifecycle phases.
type AppLifecycleEvent interface {
	AppEvent

	// OnStarting is called before the application starts.
	// This is the ideal place to perform pre-start tasks such as:
	//
	// - Validating application configuration
	// - Initializing external connections (database, cache, etc.)
	// - Setting up monitoring and health checks
	// - Performing pre-flight checks
	// - Registering additional middleware or routes
	//
	// Returns:
	//   - true: Prevent the application from starting (abort startup)
	//   - false: Allow the application to start normally
	OnStarting() bool

	// OnStarted is called after the application has started successfully.
	// This is the ideal place to perform post-start tasks such as:
	//
	// - Starting background workers or schedulers
	// - Sending startup notifications
	// - Performing warm-up operations
	// - Logging startup completion status
	// - Triggering external system notifications
	//
	// Note: At this point, the HTTP server is running and ready to accept requests.
	OnStarted()

	// OnStopping is called before the application begins its shutdown process.
	// This is the ideal place to perform pre-stop tasks such as:
	//
	// - Stopping background workers gracefully
	// - Draining request queues
	// - Notifying external systems of impending shutdown
	// - Preparing for graceful resource cleanup
	// - Setting maintenance mode flags
	OnStopping()

	// OnStopped is called after the application has completed its shutdown process.
	// This is the ideal place to perform final cleanup tasks such as:
	//
	// - Closing database connections
	// - Releasing file handles and network resources
	// - Flushing logs and metrics
	// - Sending shutdown completion notifications
	// - Performing final cleanup operations
	//
	// Note: This is the last opportunity to perform cleanup before process termination.
	OnStopped()
}

// ConfigAfterLoadEvent handles events after configuration loading completes.
// Implement this interface to perform post-loading tasks such as validation or derived value calculation.
type ConfigAfterLoadEvent interface {
	AppEvent
	// OnConfigAfterLoad is called after all configuration files have been successfully loaded.
	OnConfigAfterLoad(configure config.Configure)
}

// ContainerRefreshBeforeEvent handles events triggered before container refresh.
// Implement this interface to perform pre-refresh setup and configuration tasks.
type ContainerRefreshBeforeEvent interface {
	AppEvent
	// OnContainerRefreshBefore is called before the IoC container begins its refresh process.
	// This is the ideal place to perform container preparation tasks such as:
	//
	// - Registering additional bean definitions
	// - Registering pre-configured bean instances
	// - Modifying container configuration
	// - Performing pre-refresh validations
	// - Setting up custom bean processors
	OnContainerRefreshBefore(c *ioc.Container)
}

// ContainerRefreshAfterEvent handles events triggered after container refresh completion.
// Implement this interface to perform post-refresh finalization and validation tasks.
type ContainerRefreshAfterEvent interface {
	AppEvent
	// OnContainerRefreshAfter is called after the IoC container has completed its refresh process.
	// At this point, all beans have been created, dependencies injected, and the container is ready for use.
	// This is the ideal place to perform post-refresh tasks such as:
	//
	// - Validating bean initialization results
	// - Performing cross-bean validations
	// - Starting background services that depend on beans
	// - Logging container status and statistics
	// - Triggering application-specific initialization logic
	//
	// Note: The container state should not be modified at this point.
	OnContainerRefreshAfter(c *ioc.Container)
}
