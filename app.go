package gin_plus

import (
	"context"
	"errors"
	"fmt"
	"os"
	"os/signal"
	"sync/atomic"
	"syscall"
	"time"

	"github.com/archine/gin-plus/v4/component/config"
	"github.com/archine/gin-plus/v4/component/gplog"
	"github.com/archine/gin-plus/v4/component/gplog/zapper"
	"github.com/archine/gin-plus/v4/component/ioc"
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

// shutdownSignalBeanName is the bean name of the ShutdownSignal; inject it by type instead.
const shutdownSignalBeanName = "gin-plus.shutdownSignal"

type App struct {
	used             atomic.Bool // Guards Run: an App is single-use because the container refreshes once
	running          atomic.Bool // True while Run is in progress
	eventManager     *eventManager
	server           *server.GinServer
	confProviderFunc func() config.Provider
	loggerFunc       func(cp config.Provider) gplog.Logger
	banner           string
	shutdownSignal   *shutdownSignal
}

// New creates a new App; configure it with With before calling Run.
func New() *App {
	return &App{
		eventManager:   newEventManager(),
		server:         server.NewGinServer(),
		shutdownSignal: newShutdownSignal(),
	}
}

// With adds options to the App instance.
func (a *App) With(opts ...Option) *App {
	for _, opt := range opts {
		opt(a)
	}
	return a
}

// IsRunning reports whether Run is in progress.
func (a *App) IsRunning() bool {
	return a.running.Load()
}

// Run starts the application in the specified mode and blocks until it finishes.
// In ServerMode that is after a graceful shutdown triggered by SIGINT / SIGTERM.
//
// Failures are handled by the framework, callers never need to deal with start-up errors:
//   - before the logger is installed (invalid mode, config or logger factory), Run panics
//     with the cause, since there is nothing to log it with;
//   - afterwards, the cause is logged and the process exits with status 1.
//     Resources acquired before the failure are released first (see shutdown).
//
// An App can only be run once; subsequent calls log a warning and return immediately.
func (a *App) Run(mode RunMode) {
	if mode < InitializeMode || mode > ServerMode {
		panic(fmt.Sprintf("gin-plus: unknown run mode %d", mode))
	}
	if !a.used.CompareAndSwap(false, true) {
		gplog.Warn("Application is already running or has run, ignore this call")
		return
	}
	a.running.Store(true)
	defer a.running.Store(false)

	printBanner(a.banner)
	a.banner = ""

	a.initialize()
	if mode == InitializeMode {
		return
	}

	err := a.refreshContainer()
	if err == nil && mode == ServerMode {
		err = a.serve()
	}
	if err != nil {
		gplog.Error("Application run failed: " + err.Error())
		// os.Exit skips deferred calls; that is fine because serve has already released its resources.
		os.Exit(1)
	}
}

// initialize loads the configuration and installs the logger.
// It panics on failure because the logger is not available yet.
func (a *App) initialize() {
	if a.confProviderFunc == nil {
		a.confProviderFunc = config.NewFileProvider
	}
	cp := a.confProviderFunc()
	if cp == nil {
		panic("gin-plus: config provider function returned nil")
	}
	sysconf.Provider = cp
	a.eventManager.triggerConfigLoaded(cp)

	if a.loggerFunc == nil {
		a.loggerFunc = func(cp config.Provider) gplog.Logger {
			return zapper.NewLogger(cp)
		}
	}
	logger := a.loggerFunc(cp)
	if logger == nil {
		panic("gin-plus: logger function returned nil")
	}
	gplog.SetLogger(logger)

	// Release the factories, they are never used again.
	a.confProviderFunc, a.loggerFunc = nil, nil

	gplog.Info("Application configuration and logger initialized")
}

// refreshContainer builds the bean graph.
func (a *App) refreshContainer() error {
	// Framework-owned beans; the shutdown signal is injectable through the ShutdownSignal interface.
	ioc.RegisterBean(shutdownSignalBeanName, a.shutdownSignal)
	a.eventManager.triggerContainerRefreshBefore()

	// Drain closes the registry, so late ioc.RegisterBeanDef calls fail with a clear message.
	if err := sysctr.Container.Refresh(sysconf.Provider, sysctr.BeanRegistry.Drain()); err != nil {
		return err
	}

	gplog.Info("Bean container refreshed successfully")
	a.eventManager.triggerContainerRefreshAfter()
	return nil
}

// serve starts the HTTP server and blocks until a shutdown signal or a fatal serving error.
func (a *App) serve() error {
	if err := a.server.Init(sysconf.Provider); err != nil {
		return err
	}

	// Subscribe before starting, so a signal arriving during start-up leads to a graceful
	// shutdown instead of killing the process with half-initialized resources.
	sigCtx, stop := signal.NotifyContext(context.Background(), syscall.SIGINT, syscall.SIGTERM)
	defer stop()

	name := a.server.Name()
	gplog.Info(fmt.Sprintf("Starting %s using Gin-Engine on %s with PID %d", name, a.server.Address(), os.Getpid()))
	a.eventManager.triggerOnStarting()

	startTime := time.Now()
	serveErr, err := a.server.Start()
	if err != nil {
		// OnStarting may already have acquired resources; release them before reporting.
		return errors.Join(fmt.Errorf("start %s: %w", name, err), a.shutdown())
	}

	a.eventManager.triggerOnStarted()
	gplog.Info(fmt.Sprintf("Started %s in %v", name, time.Since(startTime)))

	var runErr error
	select {
	case <-sigCtx.Done():
		gplog.Info("Received shutdown signal, starting graceful shutdown...")
	case err = <-serveErr:
		runErr = fmt.Errorf("%s stopped unexpectedly: %w", name, err)
	}
	// Restore default signal handling: a second Ctrl+C now force-quits a stuck shutdown.
	stop()

	return errors.Join(runErr, a.shutdown())
}

// shutdown releases everything in dependency order:
//  1. close the ShutdownSignal, so long-lived handlers (SSE, websockets) can return;
//  2. stop the HTTP server gracefully, waiting for in-flight requests;
//  3. run OnStopped callbacks within the exit-delay budget, even if step 2 failed.
func (a *App) shutdown() error {
	a.shutdownSignal.close()

	err := a.server.Shutdown()
	if err != nil {
		gplog.Warn(err.Error())
	}

	a.eventManager.triggerOnStopped(a.server.ExitDelay())
	gplog.Info("Application shutdown completed")
	return err
}
