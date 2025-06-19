package event

// AppStartingEvent handles the application startup event.
// Implement this interface to perform tasks before the application begins serving requests.
type AppStartingEvent interface {
	AppEvent
	// OnApplicationStarting is triggered just before the application starts accepting requests.
	// Use this hook for final setup tasks that need to happen after container initialization.
	OnApplicationStarting()
}

// AppStartedEvent handles the application started event.
// Implement this interface to perform tasks after the application is fully running.
type AppStartedEvent interface {
	AppEvent
	// OnApplicationStarted is triggered after the application has successfully started.
	// Use this hook for post-startup tasks like logging, health checks, or notifications.
	OnApplicationStarted()
}

// AppStoppingEvent handles the application shutdown event.
// Implement this interface to perform cleanup tasks before the application stops.
type AppStoppingEvent interface {
	AppEvent
	// OnApplicationStopping is triggered when the application begins its shutdown sequence.
	// Use this hook for graceful shutdown tasks like closing connections, saving state, etc.
	OnApplicationStopping()
}

// AppStoppedEvent handles the application stopped event.
// Implement this interface to perform final cleanup after the application has stopped.
type AppStoppedEvent interface {
	AppEvent
	// OnApplicationStopped is triggered after the application has completely shut down.
	// Use this hook for final cleanup tasks and resource deallocation.
	OnApplicationStopped()
}
