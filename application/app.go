package application

import (
	"context"
	"errors"
	"fmt"
	"github.com/archine/gin-plus/v3/application/middleware"
	"github.com/archine/gin-plus/v3/banner"
	"github.com/archine/gin-plus/v3/exception/interceptor"
	"github.com/archine/gin-plus/v3/listener"
	"github.com/archine/gin-plus/v3/mvc"
	"github.com/archine/gin-plus/v3/plugin/logger"
	"github.com/archine/ioc"
	"github.com/gin-gonic/gin"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// App application instance
type App struct {
	e              *gin.Engine
	exitDelay      time.Duration
	interceptors   []mvc.MethodInterceptor
	ginMiddlewares []gin.HandlerFunc
	listeners      []listener.ApplicationListener
}

// New Create a clean application, you can add some gin middlewares to the engine
func New(listeners []listener.ApplicationListener, middlewares ...gin.HandlerFunc) *App {
	app := &App{
		exitDelay:      3 * time.Second,
		ginMiddlewares: middlewares,
	}
	configured := false
	for _, l := range listeners {
		if cl, ok := l.(listener.ConfigListener); ok {
			LoadApplicationConfigFile(cl)
			configured = true
			continue
		}
		app.listeners = append(app.listeners, l)
	}
	if !configured {
		LoadApplicationConfigFile(nil)
	}
	if Conf.Server.Env == Prod {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
	return app
}

// Default Create a default application with gin default logger, exception interception, and cross-domain middleware
func Default(listeners ...listener.ApplicationListener) *App {
	return New(listeners, gin.Logger(), interceptor.GlobalExceptionInterceptor, middleware.Cors())
}

// Banner Sets the project startup banner
func (a *App) Banner(b string) *App {
	banner.Banner = b
	return a
}

// Log Sets the log collector
func (a *App) Log(collector logger.AbstractLogger) *App {
	logger.Log = collector
	return a
}

// Interceptor Add a global interceptor
func (a *App) Interceptor(interceptor ...mvc.MethodInterceptor) *App {
	a.interceptors = append(a.interceptors, interceptor...)
	return a
}

// Run the main program entry
func (a *App) Run() {
	a.e = gin.New()
	server := &http.Server{
		Addr:                         fmt.Sprintf(":%d", Conf.Server.Port),
		ReadTimeout:                  Conf.Server.ReadTimeout,
		WriteTimeout:                 Conf.Server.WriteTimeout,
		DisableGeneralOptionsHandler: true,
	}
	server.Handler = a.e
	if len(a.ginMiddlewares) > 0 {
		a.e.Use(a.ginMiddlewares...)
	}
	a.e.MaxMultipartMemory = Conf.Server.MaxFileSize
	a.e.RemoveExtraSlash = true
	ioc.SetBeans(a.e)
	if banner.Banner != "" {
		fmt.Print(banner.Banner)
	}
	listener.DoPreApply(a.listeners)
	if len(a.interceptors) > 0 {
		a.e.Use(func(context *gin.Context) {
			var is []mvc.MethodInterceptor
			for _, ic := range a.interceptors {
				if ic.Predicate(context) {
					is = append(is, ic)
					ic.PreHandle(context)
				}
				if context.IsAborted() {
					return
				}
			}
			context.Next()
			for _, i := range is {
				i.PostHandle(context)
				if context.IsAborted() {
					return
				}
			}
		})
	}
	mvc.Apply(a.e, true)
	listener.DoPreStart(a.listeners)
	go func() {
		if err := server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Log.Fatal("Application start error, %s", err.Error())
		}
	}()
	logger.Log.Info("Application start success on Ports:[%d]", Conf.Server.Port)
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	logger.Log.Debug("Shutdown server ...")
	listener.DoPreStop(a.listeners)
	ctx, cancelFunc := context.WithTimeout(context.Background(), a.exitDelay)
	defer cancelFunc()
	if err := server.Shutdown(ctx); err != nil {
		logger.Log.Fatal("Server shutdown failure, %s", err.Error())
	}
	listener.DoPostStop(a.listeners)
	logger.Log.Debug("Server exiting ...")
}

// ReadConfig Read configuration
// v config struct pointer
func (a *App) ReadConfig(v any) *App {
	if err := GetConfReader().Unmarshal(v); err != nil {
		logger.Log.Fatal("read config error, %s", err.Error())
	}
	return a
}

// ReadConfigSub Read configuration
// v: config struct pointer
// sub: sub configuration key
func (a *App) ReadConfigSub(v any, sub string) *App {
	if err := GetConfReader().Sub(sub).Unmarshal(v); err != nil {
		logger.Log.Fatal("read config error, %s", err.Error())
	}
	return a
}

// ExitDelay Graceful exit time(default 3s), when reached to shut down the server and trigger PostStop().
func (a *App) ExitDelay(time time.Duration) *App {
	a.exitDelay = time
	return a
}
