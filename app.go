package gin_plus

import (
	"context"
	"errors"
	"fmt"
	"github.com/archine/gin-plus/v4/app"
	"github.com/archine/gin-plus/v4/internal/container"
	"net/http"
	"os"
	"os/signal"
	"syscall"

	"github.com/archine/gin-plus/v4/component/gplog"
	"github.com/archine/gin-plus/v4/internal/server"
	"github.com/archine/gin-plus/v4/internal/sysconf"
	"github.com/archine/gin-plus/v4/internal/syslog"
	"github.com/archine/gin-plus/v4/middleware"
	"github.com/gin-gonic/gin"
)

const (
	// StateInit indicates that the application is in the initial state.
	StateInit = 1
	// StateContainerPrepared indicates that the IoC container has been prepared.
	StateContainerPrepared = 2
	// StateRunning indicates that the application is currently running.
	// This state is set when the application has started successfully and is ready to handle requests.
	StateRunning = 4
)

type App struct {
	state        int
	eventManager *eventManager
	server       *server.GinServer
	appContext   *app.Context
	container    *container.Container
}

// New creates a new instance of the App with optional configurations.
//
// Args:
//   - opts: A variadic list of options to configure the App instance.
func New(opts ...Option) *App {
	a := &App{
		state:        StateInit,
		eventManager: &eventManager{},
		server:       &server.GinServer{},
		container:    container.NewContainer(),
	}
	a.appContext = app.NewContext(a.container)

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

// PrepareContainer initializes and refreshes the IoC container.
// This method ensures that all beans are created and injected properly.
// In most cases, you do not need to call this method explicitly, as it is automatically invoked when the application starts.
// Only call this method directly if you need to initialize the container without starting the server (e.g., for testing or tooling purposes).
func (a *App) PrepareContainer() {
	if a.state&StateContainerPrepared != 0 {
		gplog.Warn("Application container is already prepared.")
		return
	}

	a.container.Refresh()
	a.eventManager.TriggerContainerRefreshAfter(a.appContext)

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

	err := a.server.Run(a.appContext)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		gplog.Error(fmt.Sprintf("Application failed to start: %v", err))
		return
	}

	a.state |= StateRunning
	a.eventManager.TriggerOnStarted()
	gplog.Info(fmt.Sprintf("Application started successfully on [%s]", a.server.GetAddress()))

	stopSignalCh := make(chan os.Signal, 1)
	signal.Notify(stopSignalCh, syscall.SIGTERM, syscall.SIGINT)
	<-stopSignalCh

	if err = a.server.Shutdown(func(ctx context.Context) {
		a.eventManager.TriggerOnStopped(ctx)
	}); err != nil {
		gplog.Warn(fmt.Sprintf("Application failed to gracefully shutdown: %v", err))
		return
	}

	gplog.Info("Application shutdown gracefully.")
}
