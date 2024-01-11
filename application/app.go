package application

import (
	"context"
	"errors"
	"fmt"
	"github.com/archine/gin-plus/v3/banner"
	"github.com/archine/gin-plus/v3/exception"
	"github.com/archine/gin-plus/v3/mvc"
	"github.com/archine/gin-plus/v3/plugin/logger"
	"github.com/archine/ioc"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// App application instance
type App struct {
	e              *gin.Engine
	preApplyFunc   func()
	preStartFunc   func()
	preStopFunc    func()
	exitDelay      time.Duration
	interceptors   []mvc.MethodInterceptor
	ginMiddlewares []gin.HandlerFunc
	server         *http.Server
}

// New Create a clean application, you can add some gin middlewares to the engine
func New(confOptions []viper.Option, middlewares ...gin.HandlerFunc) *App {
	LoadApplicationConfigFile(confOptions)
	if Conf.Server.Env == Prod {
		gin.SetMode(gin.ReleaseMode)
	} else {
		gin.SetMode(gin.DebugMode)
	}
	return &App{
		exitDelay:      3 * time.Second,
		ginMiddlewares: middlewares,
		server: &http.Server{
			Addr:                         fmt.Sprintf(":%d", Conf.Server.Port),
			ReadTimeout:                  Conf.Server.ReadTimeout,
			WriteTimeout:                 Conf.Server.WriteTimeout,
			DisableGeneralOptionsHandler: true,
		},
	}
}

// Default Create a default application with gin default logger, exception interception, and cross-domain middleware
func Default(confOptions ...viper.Option) *App {
	return New(
		confOptions,
		gin.Logger(),
		exception.GlobalExceptionInterceptor,
		cors.New(cors.Config{
			AllowMethods:     []string{"PUT", "PATCH", "POST", "GET", "DELETE", "OPTIONS", "HEAD"},
			AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
			ExposeHeaders:    []string{"Content-Length"},
			AllowCredentials: true,
			AllowOriginFunc: func(origin string) bool {
				return true
			},
		}))
}

// Banner Sets the project startup banner
func (a *App) Banner(b string) *App {
	banner.Banner = b
	return a
}

// Log Sets the log collector
func (a *App) Log(collector logger.AbstractLogger) *App {
	collector.Init()
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
	if logger.Log == nil {
		logger.Log = &logger.DefaultLog{}
	}
	a.e = gin.New()
	a.server.Handler = a.e
	if len(a.ginMiddlewares) > 0 {
		a.e.Use(a.ginMiddlewares...)
	}
	a.e.MaxMultipartMemory = Conf.Server.MaxFileSize
	a.e.RemoveExtraSlash = true
	ioc.SetBeans(a.e)
	if banner.Banner != "" {
		fmt.Print(banner.Banner)
	}
	if a.preApplyFunc != nil {
		a.preApplyFunc()
	}
	if len(a.interceptors) > 0 {
		a.e.Use(func(context *gin.Context) {
			var is []mvc.MethodInterceptor
			for _, interceptor := range a.interceptors {
				if interceptor.Predicate(context) {
					is = append(is, interceptor)
					interceptor.PreHandle(context)
				}
				if context.IsAborted() {
					return
				}
			}
			context.Next()
			for _, interceptor := range is {
				interceptor.PostHandle(context)
				if context.IsAborted() {
					return
				}
			}
		})
	}
	mvc.Apply(a.e, true)
	if a.preStartFunc != nil {
		a.preStartFunc()
	}
	go func() {
		if err := a.server.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			logger.Log.Fatalf("Application start error, %s", err.Error())
		}
	}()
	logger.Log.Debugf("Application start success on Ports:[%d]", Conf.Server.Port)
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	logger.Log.Debug("Shutdown server ...")
	if a.preStopFunc != nil {
		a.preStopFunc()
	}
	ctx, cancelFunc := context.WithTimeout(context.Background(), a.exitDelay)
	defer cancelFunc()
	if err := a.server.Shutdown(ctx); err != nil {
		logger.Log.Fatalf("Server shutdown failure, %s", err.Error())
	}
	logger.Log.Debug("Server exiting ...")
}

// ReadConfig Read configuration
// v config struct pointer
func (a *App) ReadConfig(v any) *App {
	if err := GetConfReader().Unmarshal(v); err != nil {
		logger.Log.Fatalf("read config error, %s", err.Error())
	}
	return a
}

// PreApply triggered before mvc starts, Before the project starts.
// This is where you can provide basic services, such as set beans.
// Of course, you can also perform logic here that doesn't require obtaining beans.
func (a *App) PreApply(f func()) *App {
	if f == nil {
		logger.Log.Fatalf("apply before func cannot be null.")
	}
	a.preApplyFunc = f
	return a
}

// PreStart The last event before the project starts, dependency injection is all finished and ready to run.
// You can execute any logic here.
func (a *App) PreStart(f func()) *App {
	a.preStartFunc = f
	return a
}

// PreStop The event before the application stops can be performed here to close some resources
func (a *App) PreStop(f func()) *App {
	a.preStopFunc = f
	return a
}

// PostStop Events after the application has stopped can perform other closing operations here
func (a *App) PostStop(f func()) *App {
	a.server.RegisterOnShutdown(f)
	return a
}

// ExitDelay Graceful exit time(default 3s), when reached to shut down the server and trigger PostStop().
func (a *App) ExitDelay(time time.Duration) *App {
	a.exitDelay = time
	return a
}
