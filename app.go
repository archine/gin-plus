package gin_plus

import (
	"context"
	"errors"
	"fmt"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/archine/gin-plus/v4/component/config"
	"github.com/archine/gin-plus/v4/component/gplog"
	"github.com/archine/gin-plus/v4/component/gplog/zapper"
	"github.com/archine/gin-plus/v4/internal/server"
	"github.com/archine/gin-plus/v4/internal/vars/sysconf"
	"github.com/archine/gin-plus/v4/internal/vars/sysctr"
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
	confProviderFunc func() config.Provider
	loggerFunc       func(cp config.Provider) gplog.Logger
	initialized      atomic.Bool
	banner           string
}

// New creates a new instance of the App with optional configurations.
func New() *App {
	a := &App{
		eventManager: newEventManager(),
		server:       server.NewGinServer(),
	}

	return a
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
	if !a.state.CompareAndSwap(false, true) {
		gplog.Warn("Application is already running")
		return
	}

	printBanner(a.banner)
	a.banner = ""

	a.initialize()

	switch mode {
	case InitializeMode:
		// do nothing
	case ContainerMode:
		a.refreshContainer()
	case ServerMode:
		a.refreshContainer()
		a.startServer()
	default:
		gplog.Warn("Unknown run mode, defaulting to ServerMode")
		a.refreshContainer()
		a.startServer()
	}
}

// initialize initializes the application configuration and logger.
// This method is called automatically when the application starts.
// It's safe to call multiple times - subsequent calls are no-ops.
func (a *App) initialize() {
	// Fast path: already initialized
	if a.initialized.Load() {
		return
	}

	// Initialize config provider
	if a.confProviderFunc == nil {
		a.confProviderFunc = config.NewFileProvider
	}
	sysconf.Provider = a.confProviderFunc()
	a.eventManager.triggerConfigLoaded(sysconf.Provider)

	// Initialize logger
	if a.loggerFunc == nil {
		a.loggerFunc = func(cp config.Provider) gplog.Logger {
			return zapper.NewLogger(cp)
		}
	}
	logger := a.loggerFunc(sysconf.Provider)
	gplog.SetLogger(logger)

	// Clear function references to allow GC
	a.confProviderFunc = nil
	a.loggerFunc = nil

	a.initialized.Store(true)
	gplog.Info("Application configuration and logger initialized")
}

// refreshContainer refreshes the bean container.
func (a *App) refreshContainer() {
	a.eventManager.triggerContainerRefreshBefore()

	sysctr.Container.Refresh()
	gplog.Info("Bean container refreshed successfully")

	a.eventManager.triggerContainerRefreshAfter()
}

// startServer extracts the server startup logic for better readability
func (a *App) startServer() {
	if err := a.server.Init(); err != nil {
		gplog.Error("Starting server failed: " + err.Error())
		return
	}

	serverName := a.server.GetName()
	gplog.Info(fmt.Sprintf("Starting %s using Gin-Engine on %s with PID %d",
		serverName, a.server.GetAddress(), os.Getpid()))

	a.eventManager.triggerOnStarting()

	startTime := time.Now()
	if err := a.server.Run(); err != nil && !errors.Is(err, http.ErrServerClosed) {
		gplog.Error("Starting " + serverName + " failed: " + err.Error())
		return
	}

	a.eventManager.triggerOnStarted()
	gplog.Info(fmt.Sprintf("Started %s in %v", serverName, time.Since(startTime)))

	a.waitForShutdown()
}

// waitForShutdown handles graceful shutdown logic
func (a *App) waitForShutdown() {
	stopSignalCh := make(chan os.Signal, 1)
	signal.Notify(stopSignalCh, syscall.SIGTERM, syscall.SIGINT)
	<-stopSignalCh

	gplog.Info("Received shutdown signal, starting graceful shutdown...")

	if err := a.server.Shutdown(func(ctx context.Context) {
		a.eventManager.triggerOnStopped(ctx)
	}); err != nil {
		gplog.Warn("Graceful shutdown failed: " + err.Error())
		return
	}

	gplog.Info("Application shutdown completed successfully")
}
