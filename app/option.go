package app

import (
	"github.com/archine/gin-plus/v4/component/config"
	"github.com/archine/gin-plus/v4/component/event"
	"github.com/archine/gin-plus/v4/component/gplog"
	"github.com/archine/gin-plus/v4/internal/server"
	"github.com/archine/gin-plus/v4/internal/syslink"
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
		app.configure = confFunc()
		if app.configure == nil {
			panic("app configure is nil")
		}

		var conf *server.Config
		if app.configure != nil {
			err := app.configure.Unmarshal("gin_plus", conf)
			if err != nil {
				panic("app configure unmarshal error: " + err.Error())
			}
		}

		app.config = conf
		app.eventManager.TriggerConfigAfterLoad(app.configure)
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
func WithEvent(events ...event.AppEvent) Option {
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
		syslink.SetGlobalLogger(loggerFunc(app.configure))
	}
}
