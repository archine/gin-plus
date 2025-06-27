package app

import "github.com/archine/gin-plus/v4/component/config"

type Option func(app *App) // Option is a function that modifies the App instance.

func WithConfigure(conf config.Configure) Option {
	return func(app *App) {
		app.configure = &conf
	}
}
