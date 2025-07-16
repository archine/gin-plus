package gin_plus

import (
	"context"
	"errors"
	"fmt"
	"github.com/archine/gin-plus/v4/app"
	"github.com/archine/gin-plus/v4/component/gpconf"
	"github.com/archine/gin-plus/v4/component/gplog/gplogcore"
	"github.com/archine/gin-plus/v4/internal/container"
	"github.com/archine/gin-plus/v4/internal/vars/syscontainer"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/archine/gin-plus/v4/component/gplog"
	"github.com/archine/gin-plus/v4/internal/server"
	"github.com/archine/gin-plus/v4/middleware"
	"github.com/gin-gonic/gin"
)

const (
	// StateInit indicates that the application is in the initial state.
	StateInit = 1
	// StateContainerRefreshed indicates that the application container has been refreshed.
	StateContainerRefreshed = 2
	// StateRunning indicates that the application is currently running.
	// This state is set when the application has started successfully and is ready to handle requests.
	StateRunning = 4
)

type App struct {
	state        int
	eventManager *eventManager
	server       *server.GinServer
	appContext   *app.Context
}

// New creates a new instance of the App with optional configurations.
//
// Args:
//   - opts: A variadic list of options to configure the App instance.
func New(opts ...Option) *App {
	a := &App{
		state:        StateInit,
		eventManager: &eventManager{},
		appContext:   app.NewContext(),
		server:       server.NewGinServer(),
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
		WithConfigure(gpconf.NewLocalFileConfigure),
		WithLogger(gplogcore.NewZapLogger),
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

// RefreshContainer refreshes the bean container.
// This method ensures that all beans are created and injected properly.
// In most cases, you do not need to call this method explicitly, as it is automatically invoked when the application starts.
// Only call this method directly if you need to initialize the container without starting the server (e.g., for testing or tooling purposes).
func (a *App) RefreshContainer() {
	if a.state&StateContainerRefreshed != 0 {
		gplog.Warn("Application container is already prepared")
		return
	}

	syscontainer.Container = container.NewContainer()

	a.eventManager.TriggerContainerRefreshBefore(a.appContext)
	syscontainer.Container.Refresh()
	a.eventManager.TriggerContainerRefreshAfter(a.appContext)

	a.state |= StateContainerRefreshed
	gplog.Info("Application container has been refreshed and is ready for use")
}

// Run starts the application.
func (a *App) Run() {
	if a.state&StateRunning != 0 {
		gplog.Warn("Application is already running")
		return
	}

	if a.state&StateContainerRefreshed == 0 {
		a.RefreshContainer()
	}

	a.server.Init()
	gplog.Info(fmt.Sprintf("Starting %s using Gin-Engine on %s with PID %d", a.server.GetName(), a.server.GetAddress(), os.Getpid()))

	continueRun := a.eventManager.TriggerOnStarting()
	if !continueRun {
		return
	}

	printBanner()

	startTime := time.Now()
	err := a.server.Run(a.appContext)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		gplog.Error(fmt.Sprintf("Starting %s failed: %v", a.server.GetName(), err))
		return
	}

	a.state |= StateRunning
	a.eventManager.TriggerOnStarted()
	gplog.Info(fmt.Sprintf("Started %s in %v", a.server.GetName(), time.Since(startTime)))

	stopSignalCh := make(chan os.Signal, 1)
	signal.Notify(stopSignalCh, syscall.SIGTERM, syscall.SIGINT)
	<-stopSignalCh
	gplog.Info("Received shutdown signal, starting graceful shutdown...")

	if err = a.server.Shutdown(func(ctx context.Context) {
		a.eventManager.TriggerOnStopped(ctx)
	}); err != nil {
		gplog.Warn(fmt.Sprintf("Graceful shutdown failed: %v", err))
		return
	}

	gplog.Info("Application shutdown completed successfully")
}
