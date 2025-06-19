package event

// ContextBeforeInitEvent defines the event that is triggered before the application context initialization.
// This event is fired just before Gin routes and dependency injection are set up.
type ContextBeforeInitEvent interface {
	AppEvent
	// OnContextBeforeInit is called before the application context is initialized.
	// This includes the setup of Gin routing and dependency injection container.
	OnContextBeforeInit()
}

// ContextAfterInitEvent defines the event that is triggered after the application context initialization.
// This event is fired when all API routes have been registered and all beans have been initialized.
type ContextAfterInitEvent interface {
	AppEvent
	// OnContextAfterInit is called after the application context is fully initialized.
	// At this point, all API routes are registered and all beans are completely initialized.
	OnContextAfterInit()
}
