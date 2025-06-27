package event

import (
	"github.com/archine/gin-plus/v4/component/config"
)

// AppEvent is the base interface for all app events.
type AppEvent interface {
	// Order returns the order of the event.
	// Lower numbers indicate higher priority.
	Order() int
}

// AppStartingEvent handles the app startup event.
// Implement this interface to perform tasks before the app begins serving requests.
type AppStartingEvent interface {
	AppEvent
	// OnApplicationStarting is triggered just before the app starts accepting requests.
	// Use this hook for final setup tasks that need to happen after container initialization.
	OnApplicationStarting()
}

// AppStartedEvent handles the app started event.
// Implement this interface to perform tasks after the app is fully running.
type AppStartedEvent interface {
	AppEvent
	// OnApplicationStarted is triggered after the app has successfully started.
	// Use this hook for post-startup tasks like logging, health checks, or notifications.
	OnApplicationStarted()
}

// AppStoppingEvent handles the app shutdown event.
// Implement this interface to perform cleanup tasks before the app stops.
type AppStoppingEvent interface {
	AppEvent
	// OnApplicationStopping is triggered when the app begins its shutdown sequence.
	// Use this hook for graceful shutdown tasks like closing connections, saving state, etc.
	OnApplicationStopping()
}

// AppStoppedEvent handles the app stopped event.
// Implement this interface to perform final cleanup after the app has stopped.
type AppStoppedEvent interface {
	AppEvent
	// OnApplicationStopped is triggered after the app has completely shut down.
	// Use this hook for final cleanup tasks and resource deallocation.
	OnApplicationStopped()
}

// ConfigAfterLoadEvent handles events after configuration loading completes.
// Implement this interface to perform post-loading tasks such as validation or derived value calculation.
type ConfigAfterLoadEvent interface {
	AppEvent

	// OnConfigAfterLoad is called after all configuration files have been successfully loaded.
	OnConfigAfterLoad(configure config.Configure)
}
