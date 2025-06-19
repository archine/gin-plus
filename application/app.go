package application

import (
	"context"
	"errors"
	"fmt"
	"github.com/archine/gin-plus/v3/event"
	"github.com/archine/gin-plus/v3/internal/event_manager"
	"github.com/archine/gin-plus/v3/internal/logger"
	"github.com/archine/gin-plus/v3/module/gplog/iface"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"

	"github.com/archine/gin-plus/v3/banner"
	"github.com/archine/gin-plus/v3/internal/config"
	"github.com/archine/gin-plus/v3/module/middleware"
	"github.com/archine/gin-plus/v3/mvc"
	"github.com/gin-gonic/gin"
)

// App represents the main application structure.
type App struct {
	engine         *gin.Engine
	exitDelay      time.Duration
	ginMiddlewares []gin.HandlerFunc
	interceptors   []mvc.MethodInterceptor
	eventManager   *event_manager.AppEventManager
}

// New Create a clean application, you can add some gin middlewares to the engine
func New(middlewares ...gin.HandlerFunc) *App {
	if banner.Banner != "" {
		fmt.Println(banner.Banner)
		banner.Banner = ""
	}
	app := &App{
		exitDelay:      0,
		ginMiddlewares: middlewares,
		eventManager:   event_manager.NewEventManager(),
	}

	ioc.SetBeans(app)
	return app
}

// Default creates a default application with built-in middleware and listeners.
func Default() *App {
	return New(gin.Logger(), middleware.GlobalExceptionInterceptor)
}

// Banner sets a custom startup banner.
func (a *App) Banner(b string) *App {
	banner.Banner = b
	return a
}

// SetCustomLogger sets a custom logger for the application.
func (a *App) SetCustomLogger(customLogger iface.AbstractAppLogger) *App {
	if customLogger == nil {
		panic("custom logger cannot be nil")
	}
	logger.GlobalLogger = customLogger
	return a
}

// SetMethodInterceptor sets method interceptors for the application.
func (a *App) SetMethodInterceptor(interceptor ...mvc.MethodInterceptor) *App {
	a.interceptors = append(a.interceptors, interceptor...)
	return a
}

// SetEvents sets events for the application.
func (a *App) SetEvents(e ...event.AppEvent) *App {
	a.eventManager.Register(e...)
	return a
}

// Ready prepares the application for running.
func (a *App) Ready() {
	a.eventManager.Register(&logger.DefaultLoggerInitListener{})
	a.eventManager.Sort()

	config.Init(a.eventManager)
}

// Run starts the application server.
func (a *App) Run() {
	a.Ready()

	if config.Conf.Server.Env == config.Prod {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	a.engine = gin.New()
	server := &http.Server{
		Addr:                         fmt.Sprintf(":%d", config.Conf.Server.Port),
		ReadTimeout:                  config.Conf.Server.ReadTimeout,
		WriteTimeout:                 config.Conf.Server.WriteTimeout,
		ReadHeaderTimeout:            config.Conf.Server.ReadHeaderTimeout,
		IdleTimeout:                  config.Conf.Server.IdleTimeout,
		DisableGeneralOptionsHandler: config.Conf.Server.DisableGeneralOptions,
		Handler:                      a.engine,
	}
	if config.Conf.Server.AllowedCors {
		a.engine.Use(middleware.Cors())
	}
	if len(a.ginMiddlewares) > 0 {
		a.engine.Use(a.ginMiddlewares...)
	}

	a.engine.MaxMultipartMemory = config.Conf.Server.MaxMultipartMemory
	a.engine.RemoveExtraSlash = true
	ioc.SetBeans(a.engine)

	if len(a.interceptors) > 0 {
		a.engine.Use(func(ctx *gin.Context) {
			var appliedInterceptors []mvc.MethodInterceptor
			for _, interceptor := range a.interceptors {
				if interceptor.Predicate(ctx) {
					appliedInterceptors = append(appliedInterceptors, interceptor)
					interceptor.PreHandle(ctx)
				}
				if ctx.IsAborted() {
					return
				}
			}
			ctx.Next()
			for _, interceptor := range appliedInterceptors {
				interceptor.PostHandle(ctx)
				if ctx.IsAborted() {
					return
				}
			}
		})
	}

	mvc.Apply(a.engine, a.eventManager)

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			os.Exit(1)
		}
	}()

	time.Sleep(50 * time.Millisecond)

	quit := make(chan os.Signal, 1)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit

	var ctx context.Context
	if config.Conf.Server.ShutdownTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(context.Background(), config.Conf.Server.ShutdownTimeout)
		defer cancel()
	} else {
		ctx = context.Background()
	}

	if err := server.Shutdown(ctx); err != nil {
		os.Exit(1)
	}

	if a.exitDelay > 0 {
		time.Sleep(a.exitDelay)
	}

}

// ReadConfig loads the configuration into the provided structure.
func (a *App) ReadConfig(v any) *App {
	if err := GetConfReader().Unmarshal(v); err != nil {
		os.Exit(1)
	}
	return a
}

// ReadConfigSub loads the sub-configuration into the provided structure.
func (a *App) ReadConfigSub(key string, confPointer any) *App {
	if err := GetConfReader().UnmarshalKey(key, confPointer); err != nil {
		os.Exit(1)
	}
	return a
}

// ShutdownWaitDelay sets the delay after the server has shut down
// (default is 0 seconds). This delay allows time for any post-shutdown
// tasks (such as cleanup, logging, etc.) to complete before the process exits.
func (a *App) ShutdownWaitDelay(duration time.Duration) *App {
	a.exitDelay = duration
	return a
}
