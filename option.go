package gin_plus

import (
	"github.com/archine/gin-plus/v4/component/gpconf"
	"github.com/archine/gin-plus/v4/component/gplog/gplogcore"
	"github.com/gin-gonic/gin"
)

// Option defines a function type for configuring App instances.
// This follows the functional options pattern, allowing flexible and extensible
// configuration during App creation via the New() function.
type Option func(app *App)

// WithConfigure sets the application's configuration provider.
// If not specified, the default local file configuration will be used.
// Note: Calling this multiple times will overwrite the previous provider.
func WithConfigure(confFunc func() gpconf.Configure) Option {
	return func(app *App) {
		app.configureFunc = confFunc
	}
}

// WithLogger sets a custom logger for the application.
// The logger function receives the application configuration and should return a configured logger instance.
// This allows the logger to be configured based on the loaded configuration settings.
//
// Note: Calling this multiple times will overwrite the previous provider.
func WithLogger(loggerFunc func(conf gpconf.Configure) gplogcore.Logger) Option {
	return func(app *App) {
		app.loggerFunc = loggerFunc
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
		app.eventManager.register(events...)
	}
}

// WithBanner sets a custom banner for the application.
func WithBanner(banner string) Option {
	return func(app *App) {
		sysBanner = banner
	}
}
