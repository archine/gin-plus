package gin_plus

import (
	"github.com/archine/gin-plus/v4/component/config"
	"github.com/archine/gin-plus/v4/component/gplog"
	"github.com/gin-gonic/gin"
)

// Option defines a function type for configuring App instances.
// This follows the functional options pattern, allowing flexible and extensible
// configuration during App creation via the New() function.
type Option func(app *App)

// WithConfigProvider sets the application's configuration provider.
// If not specified, the default provider will be used.
// Note: Calling this multiple times will overwrite the previous provider.
func WithConfigProvider(providerFunc func() config.Provider) Option {
	return func(app *App) {
		app.confProviderFunc = providerFunc
	}
}

// WithLogger sets a custom log for the application.
// The log function receives the application configuration and should return a configured log instance.
// This allows the log to be configured based on the loaded configuration settings.
//
// Note: Calling this multiple times will overwrite the previous provider.
func WithLogger(loggerFunc func(cp config.Provider) gplog.Logger) Option {
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
