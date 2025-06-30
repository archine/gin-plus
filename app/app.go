package app

import (
	"github.com/archine/gin-plus/v4/internal/sysconf"
	"github.com/archine/gin-plus/v4/internal/syslink"
	"github.com/archine/gin-plus/v4/middleware"

	"github.com/archine/gin-plus/v4/component/config"
	"github.com/archine/gin-plus/v4/component/gplog"
	"github.com/archine/gin-plus/v4/internal/sysevent"
	"github.com/archine/gin-plus/v4/internal/syslog"
	"github.com/gin-gonic/gin"
)

const (
	// Application states
	// These constants represent the different states of the application lifecycle.

	// StateInit indicates that the application is in the initial state.
	StateInit = 1
	// StateContainerPrepared indicates that the IoC container has been prepared.
	StateContainerPrepared = 2
	// StateRunning indicates that the application is currently running.
	// This state is set when the application has started successfully and is ready to handle requests.
	StateRunning = 4
	// StateStopped indicates that the application has been stopped.
	// This state is set when the application has been gracefully shut down.
	// It can be used to check if the application is still active or has completed its lifecycle
	StateStopped = 8
)

type App struct {
	state        int
	middlewares  []gin.HandlerFunc
	eventManager *sysevent.Manager
	configure    config.Configure
}

// New creates a new instance of the App with optional configurations.
//
// Args:
//   - opts: A variadic list of options to configure the App instance.
func New(opts ...Option) *App {
	a := &App{
		state:        StateInit,
		eventManager: sysevent.NewEventManager(),
	}

	for _, opt := range opts {
		opt(a)
	}

	return a
}

// Default creates a new App instance with default configurations.
// This function sets up the application with a local file configuration,
// a default logger, and some global middlewares.
func Default() *App {
	return New(
		WithConfigure(sysconf.NewLocalFileConfigure),
		WithLogger(syslog.NewZapLogger),
		WithMiddleware(middleware.GlobalExceptionInterceptor, gin.Logger()),
	)
}

// With adds options to the App instance.
// This method allows you to configure the App instance with various options.
func (a *App) With(opts ...Option) *App {
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// PrepareContainer prepares the IoC container.
// This method initializes the container and triggers the refresh process.
// It is called after the basic preparation to ensure that all beans are properly created and injected.
func (a *App) PrepareContainer() {
	if a.state&StateContainerPrepared != 0 {
		gplog.Warn("Application container is already prepared.")
		return
	}

	ct := syslink.GetContainer()

	a.eventManager.TriggerContainerRefreshBefore(ct)
	syslink.RefreshContainer()
	a.eventManager.TriggerContainerRefreshAfter(syslink.GetContainer())

	a.state |= StateContainerPrepared
	gplog.Info("Application container prepared with refresh completed.")
}

// Run starts the application.
func (a *App) Run() {
	if a.state&StateRunning != 0 {
		gplog.Warn("Application is already running.")
		return
	}

	printBanner()

	if a.state&StateContainerPrepared == 0 {
		a.PrepareContainer()
	}

	a.state |= StateRunning
	gplog.Info("Application is now running.")

	// Here you would typically start the HTTP server or any other main loop.
	// For example:
	// ginEngine := gin.Default()
	// ginEngine.Run(":8080")
}
