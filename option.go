package gin_plus

import (
	"github.com/archine/gin-plus/v4/component/config"
	"github.com/archine/gin-plus/v4/component/gplog"
	"github.com/archine/gin-plus/v4/internal/vars/sysconf"
	"github.com/gin-gonic/gin"
)

// Option defines a function type for configuring App instances.
// This follows the functional options pattern, allowing flexible and extensible
// configuration during App creation via the New() function.
type Option func(app *App)

// WithConfigure sets the configuration provider for the application.
// If not provided, the app will use a default local file configuration.
func WithConfigure(confFunc func() config.Configure) Option {
	return func(app *App) {
		cf := confFunc()
		if cf == nil {
			panic("configuration provider is nil, please use WithConfigure() to set a configuration provider")
		}
		sysconf.ProjectConfigure = cf
		app.eventManager.TriggerConfigAfterLoad(cf)
	}
}

// WithMiddleware registers Gin middlewares to the application.
// Middlewares are executed in the order they are added.
func WithMiddleware(middlewares ...gin.HandlerFunc) Option {
	return func(app *App) {
		app.server.RegisterMiddleware(middlewares...)
	}
}

// WithEvent registers application lifecycle events.
// Events are managed by the event manager and triggered during app lifecycle.
func WithEvent(events ...Event) Option {
	return func(app *App) {
		app.eventManager.Register(events...)
	}
}

// WithBanner sets a custom banner for the application.
func WithBanner(banner string) Option {
	return func(app *App) {
		sysBanner = banner
	}
}

// WithLogger sets a custom logger for the application.
// The logger function receives the application configuration and should return a configured logger instance.
// This allows the logger to be configured based on the loaded configuration settings.
//
// Note: This option should be used after WithConfigure() to ensure configuration is available.
// The logger will be set as the global logger for the entire application.
func WithLogger(loggerFunc func(conf config.Configure) gplog.Logger) Option {
	return func(app *App) {
		logger := loggerFunc(app.appContext.GetConfigure())
		if logger == nil {
			panic("logger is nil, please ensure the logger function returns a valid logger instance")
		}
		gplog.Set(logger)
	}
}
