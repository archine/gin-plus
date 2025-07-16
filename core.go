package gin_plus

import (
	"context"
	"errors"
	"fmt"
	"github.com/archine/gin-plus/v4/app"
	"github.com/archine/gin-plus/v4/component/gpconf"
	"github.com/archine/gin-plus/v4/component/gplog/gplogcore"
	"github.com/archine/gin-plus/v4/internal/vars/sysconf"
	"github.com/archine/gin-plus/v4/internal/vars/syscontainer"
	"github.com/archine/gin-plus/v4/internal/vars/syslog"
	"net/http"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/archine/gin-plus/v4/component/gplog"
	"github.com/archine/gin-plus/v4/internal/server"
	"github.com/archine/gin-plus/v4/middleware"
	"github.com/gin-gonic/gin"
)

// runMode defines the application's running mode.
type runMode int

const (
	// ConfigMode - In this mode, the application initializes with configuration loading and logger setup, but does not start the HTTP server.
	ConfigMode runMode = iota

	// ContainerMode - Extends ConfigMode by creating and injecting dependencies into the container, while still keeping the HTTP server inactive.
	ContainerMode

	// ServerMode - Fully operational mode where the application starts the HTTP server after completing all initializations.
	ServerMode
)

type App struct {
	state         atomic.Bool
	eventManager  *eventManager
	server        *server.GinServer
	appContext    app.ApplicationContext
	configureFunc func() gpconf.Configure
	loggerFunc    func(conf gpconf.Configure) gplogcore.Logger
}

// New creates a new instance of the App with optional configurations.
//
// Args:
//   - opts: A variadic list of options to configure the App instance.
func New(opts ...Option) *App {
	a := &App{
		eventManager: newEventManager(),
		appContext:   newSysContext(),
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

// Run starts the application.
// Args:
//   - mode: The run mode of the application, either ConfigMode, ContainerMode, or ServerMode.
func (a *App) Run(mode runMode) {
	if a.state.Load() {
		gplog.Warn("Application is already running")
		return
	}

	sysconf.InitConfigure(a.configureFunc)
	a.eventManager.triggerConfigAfterLoad(sysconf.ProjectConfigure)

	syslog.InitializeLogger(a.loggerFunc)

	if mode == ConfigMode {
		gplog.Info("Started in ConfigMode, configuration loaded and logger initialized")
		return
	}

	a.refreshContainer()

	if mode == ContainerMode {
		gplog.Info("Started in ContainerMode, container initialized with dependencies")
		return
	}

	a.server.Init()
	gplog.Info(fmt.Sprintf("Starting %s using Gin-Engine on %s with PID %d", a.server.GetName(), a.server.GetAddress(), os.Getpid()))

	continueRun := a.eventManager.triggerOnStarting()
	if !continueRun {
		gplog.Warn("Application startup aborted by event handlers")
		return
	}

	printBanner()

	startTime := time.Now()
	err := a.server.Run(a.appContext)
	if err != nil && !errors.Is(err, http.ErrServerClosed) {
		gplog.Error(fmt.Sprintf("Starting %s failed: %v", a.server.GetName(), err))
		return
	}

	a.state.Store(true)
	a.eventManager.triggerOnStarted()
	gplog.Info(fmt.Sprintf("Started %s in %v", a.server.GetName(), time.Since(startTime)))

	stopSignalCh := make(chan os.Signal, 1)
	signal.Notify(stopSignalCh, syscall.SIGTERM, syscall.SIGINT)
	<-stopSignalCh
	gplog.Info("Received shutdown signal, starting graceful shutdown...")

	if err = a.server.Shutdown(func(ctx context.Context) {
		a.eventManager.triggerOnStopped(ctx)
	}); err != nil {
		gplog.Warn(fmt.Sprintf("Graceful shutdown failed: %v", err))
		return
	}

	gplog.Info("Application shutdown completed successfully")
}

// RefreshContainer refreshes the bean container.
// This method ensures that all beans are created and injected properly.
// In most cases, you do not need to call this method explicitly, as it is automatically invoked when the application starts.
// Only call this method directly if you need to initialize the container without starting the server (e.g., for testing or tooling purposes).
func (a *App) refreshContainer() {
	syscontainer.Initialize()

	a.eventManager.triggerContainerRefreshBefore(a.appContext)
	syscontainer.Container.Refresh()
	a.eventManager.triggerContainerRefreshAfter(a.appContext)

	gplog.Info("Application container has been refreshed and is ready for use")
}
