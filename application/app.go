package application

import (
	"context"
	"errors"
	"fmt"
	"github.com/archine/gin-plus/v3/banner"
	"github.com/archine/gin-plus/v3/internal"
	"github.com/archine/gin-plus/v3/internal/config"
	"github.com/archine/gin-plus/v3/listener"
	"github.com/archine/gin-plus/v3/module/middleware"
	"github.com/archine/gin-plus/v3/mvc"
	"github.com/archine/ioc"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// App represents the main application structure.
type App struct {
	engine         *gin.Engine
	exitDelay      time.Duration
	interceptors   []mvc.MethodInterceptor
	ginMiddlewares []gin.HandlerFunc
	listeners      []listener.ApplicationListener
}

// New Create a clean application, you can add some gin middlewares to the engine
func New(listeners []listener.ApplicationListener, middlewares ...gin.HandlerFunc) *App {
	if banner.Banner != "" {
		fmt.Println(banner.Banner)
		banner.Banner = ""
	}
	app := &App{
		exitDelay:      3 * time.Second,
		ginMiddlewares: middlewares,
	}

	var configLoaded bool
	for _, l := range listeners {
		if cl, ok := l.(listener.ConfigListener); ok {
			config.LoadByCommand(cl)
			configLoaded = true
			continue
		}
		app.listeners = append(app.listeners, l)
	}
	if !configLoaded {
		config.LoadByCommand(nil)
	}
	if config.Conf.Server.Env == config.Prod {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}

	return app
}

// Default creates a default application with built-in middleware and listeners.
func Default(listeners ...listener.ApplicationListener) *App {
	return New(listeners, gin.Logger(), middleware.GlobalExceptionInterceptor)
}

// Banner sets a custom startup banner.
func (a *App) Banner(b string) *App {
	banner.Banner = b
	return a
}

// Interceptor Adds a global interceptor
func (a *App) Interceptor(interceptor ...mvc.MethodInterceptor) *App {
	a.interceptors = append(a.interceptors, interceptor...)
	return a
}

// Run starts the application server.
func (a *App) Run() {
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

	if len(a.ginMiddlewares) > 0 {
		a.engine.Use(a.ginMiddlewares...)
		if config.Conf.Server.AllowedCors {
			a.engine.Use(middleware.Cors())
		}
	}

	internal.Logger.Info("Gin middlewares loaded.")
	a.engine.MaxMultipartMemory = config.Conf.Server.MaxMultipartMemory
	a.engine.RemoveExtraSlash = true
	ioc.SetBeans(a.engine)

	listener.DoPreApply(a.listeners)

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

	mvc.Apply(a.engine, true)
	internal.Logger.Info("API application setup complete.")
	listener.DoPreStart(a.listeners)

	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			internal.Logger.Fatal("Application startup failed: %s", err.Error())
		}
	}()

	time.Sleep(100 * time.Millisecond)
	internal.Logger.Info("Application started successfully on port: %d", config.Conf.Server.Port)

	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	internal.Logger.Info("Shutting down server...")

	listener.DoPreStop(a.listeners)

	ctx, cancel := context.WithTimeout(context.Background(), a.exitDelay)
	defer cancel()
	if err := server.Shutdown(ctx); err != nil {
		internal.Logger.Fatal("Server shutdown failed: %s", err.Error())
	}

	listener.DoPostStop(a.listeners)
	internal.Logger.Info("Server exited.")
}

// ReadConfig loads the configuration into the provided structure.
func (a *App) ReadConfig(v any) *App {
	if err := GetConfReader().Unmarshal(v); err != nil {
		internal.Logger.Fatal("Failed to read config, %s", err.Error())
	}
	return a
}

// ReadConfigSub loads the sub-configuration into the provided structure.
func (a *App) ReadConfigSub(v any, sub string) *App {
	if err := GetConfReader().Sub(sub).Unmarshal(v); err != nil {
		internal.Logger.Fatal("Failed to read sub-config, %s", err.Error())
	}
	return a
}

// ExitDelay sets the delay for a graceful shutdown (default is 3 seconds).
// This delay is the time given to the server to complete active requests before shutting down.
func (a *App) ExitDelay(duration time.Duration) *App {
	a.exitDelay = duration
	return a
}
