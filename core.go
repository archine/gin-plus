package gin_plus

import (
	"context"
	"errors"
	"fmt"
	"github.com/archine/gin-plus/v4/app"
	"github.com/archine/gin-plus/v4/component/config"
	"github.com/archine/gin-plus/v4/component/log/logcore"
	"github.com/archine/gin-plus/v4/internal/vars/sysconf"
	"github.com/archine/gin-plus/v4/internal/vars/sysctr"
	"github.com/archine/gin-plus/v4/internal/vars/syslog"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/archine/gin-plus/v4/component/log"
	"github.com/archine/gin-plus/v4/internal/server"
	"github.com/archine/gin-plus/v4/middleware"
	"github.com/gin-gonic/gin"
)

// RunMode defines the application's running mode.
type RunMode int

const (
	// InitializeMode initializes configuration and logger only.
	// Useful for configuration validation or testing.
	InitializeMode RunMode = iota

	// ContainerMode extends InitializeMode by creating and injecting dependencies
	// into the container, while keeping the HTTP server inactive.
	ContainerMode

	// ServerMode fully operational mode where the application starts the HTTP server
	// after completing all initializations.
	ServerMode
)

type App struct {
	state            atomic.Bool
	eventManager     *eventManager
	server           *server.GinServer
	appContext       app.ApplicationContext
	confProviderFunc func() config.Provider
	loggerFunc       func(cp config.Provider) logcore.Logger
}

// New creates a new instance of the App with optional configurations.
func New() *App {
	a := &App{
		eventManager: newEventManager(),
		appContext:   newSysContext(),
		server:       server.NewGinServer(),
	}

	return a
}

// Default creates a new App instance with default configurations.
// This function sets up the application with a local file configuration,
// a default log, and some global middlewares.
func Default() *App {
	return New().With(
		WithMiddleware(
			middleware.GlobalExceptionInterceptor,
			gin.Logger(),
		),
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

// IsRunning returns whether the application is currently running
func (a *App) IsRunning() bool {
	return a.state.Load()
}

// Run starts the application in the specified mode.
// Args:
//   - mode: The run mode of the application (InitializeMode, ContainerMode, or ServerMode).
func (a *App) Run(mode RunMode) {
	if a.state.Load() {
		log.Warn("Application is already running")
		return
	}

	a.state.Store(true)
	printBanner()
	a.initialize()

	switch mode {
	case InitializeMode:
		log.Info("Application started in [InitializeMode] - configuration and logger initialized")
	case ContainerMode:
		a.refreshContainer()
		log.Info("Application started in [ContainerMode] - IoC container initialized and ready")
	case ServerMode:
		a.refreshContainer()
		a.startServer()
	default:
		log.Warn("Unknown run mode, defaulting to ServerMode")
		a.refreshContainer()
		a.startServer()
	}
}

// initialize initializes the application configuration and logger.
// This method is called automatically when the application starts.
func (a *App) initialize() {
	if a.confProviderFunc == nil {
		a.confProviderFunc = config.NewFileProvider
	}
	sysconf.Provider = a.confProviderFunc()
	a.eventManager.triggerConfigLoaded(sysconf.Provider)

	if a.loggerFunc == nil {
		a.loggerFunc = logcore.NewZapLogger
	}
	syslog.Log = a.loggerFunc(sysconf.Provider)

	a.confProviderFunc = nil
	a.loggerFunc = nil
	log.Info("Application configuration and logger initialized")
}

// refreshContainer refreshes the bean container.
func (a *App) refreshContainer() {
	a.eventManager.triggerContainerRefreshBefore(a.appContext)
	sysctr.Container.Refresh()
	a.eventManager.triggerContainerRefreshAfter(a.appContext)

	log.Info("Application container refreshed and ready")
}

// startServer extracts the server startup logic for better readability
func (a *App) startServer() {
	err := a.server.Init()
	if err != nil {
		log.Error(fmt.Sprintf("Starting server failed: %v", err))
		return
	}
	log.Info(fmt.Sprintf("Starting %s using Gin-Engine on %s with PID %d", a.server.GetName(), a.server.GetAddress(), os.Getpid()))

	a.eventManager.triggerOnStarting()

	startTime := time.Now()
	if err = a.server.Run(a.appContext); err != nil && !errors.Is(err, http.ErrServerClosed) {
		log.Error(fmt.Sprintf("Starting %s failed: %v", a.server.GetName(), err))
		return
	}

	a.eventManager.triggerOnStarted()
	log.Info(fmt.Sprintf("Started %s in %v", a.server.GetName(), time.Since(startTime)))

	a.waitForShutdown()
}

// waitForShutdown handles graceful shutdown logic
func (a *App) waitForShutdown() {
	stopSignalCh := make(chan os.Signal, 1)
	signal.Notify(stopSignalCh, syscall.SIGTERM, syscall.SIGINT)
	<-stopSignalCh

	log.Info("Received shutdown signal, starting graceful shutdown...")

	if err := a.server.Shutdown(func(ctx context.Context) {
		a.eventManager.triggerOnStopped(ctx)
	}); err != nil {
		log.Warn(fmt.Sprintf("Graceful shutdown failed: %v", err))
		return
	}

	log.Info("Application shutdown completed successfully")
}
