package application

import (
	"context"
	"errors"
	"fmt"
	"github.com/archine/gin-plus/v3/banner"
	"github.com/archine/gin-plus/v3/exception"
	"github.com/archine/gin-plus/v3/mvc"
	"github.com/archine/gin-plus/v3/plugin"
	"github.com/archine/ioc"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
	"os"
	"os/signal"
	"syscall"
	"time"
)

// App application instance
type App struct {
	e            *gin.Engine
	configReader *viper.Viper
	preApplyFunc func()
	preStartFunc func()
	preStopFunc  func()
	postStopFunc func()
	banner       string
	exitDelay    time.Duration
}

// New Create a clean application, you can add some gin middlewares to the engine
func New(confReaderOptions []viper.Option, middlewares ...gin.HandlerFunc) *App {
	configReader := LoadApplicationConfigFile(confReaderOptions)
	plugin.InitLog(Env.LogLevel)
	if Env.LogLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	if len(middlewares) > 0 {
		engine.Use(middlewares...)
	}
	engine.MaxMultipartMemory = Env.MaxFileSize
	engine.RemoveExtraSlash = true
	ioc.SetBeans(engine)
	return &App{
		e:            engine,
		configReader: configReader,
		banner:       banner.Banner,
		exitDelay:    3 * time.Second,
	}
}

// Default Create a default application with log printing, exception interception, and cross-domain middleware
func Default(confReaderOptions ...viper.Option) *App {
	configReader := LoadApplicationConfigFile(confReaderOptions)
	plugin.InitLog(Env.LogLevel)
	if Env.LogLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	engine.Use(plugin.LogMiddleware())
	engine.Use(exception.GlobalExceptionInterceptor)
	engine.Use(cors.New(cors.Config{
		AllowMethods:     []string{"PUT", "PATCH", "POST", "GET", "DELETE", "OPTIONS", "HEAD"},
		AllowHeaders:     []string{"Origin", "Authorization", "Content-Type"},
		ExposeHeaders:    []string{"Content-Length"},
		AllowCredentials: true,
		AllowOriginFunc: func(origin string) bool {
			return true
		},
	}))
	engine.MaxMultipartMemory = Env.MaxFileSize
	engine.RemoveExtraSlash = true
	ioc.SetBeans(engine)
	return &App{
		e:            engine,
		configReader: configReader,
		banner:       banner.Banner,
		exitDelay:    3 * time.Second,
	}
}

// Banner Sets the project startup banner
func (a *App) Banner(banner string) *App {
	a.banner = banner
	return a
}

// Run the main program entry
// interceptors: link{mvc.MethodInterceptor}
func (a *App) Run(interceptors ...mvc.MethodInterceptor) {
	if a.banner != "" {
		fmt.Print(a.banner)
	}
	if a.preApplyFunc != nil {
		a.preApplyFunc()
	}
	if len(interceptors) > 0 {
		a.e.Use(func(context *gin.Context) {
			is := interceptors
			for _, interceptor := range is {
				if interceptor.Predicate(context) {
					interceptor.PreHandle(context)
					if context.IsAborted() {
						break
					}
					context.Next()
					interceptor.PostHandle(context)
					if context.IsAborted() {
						break
					}
				}
			}
		})
	}
	mvc.Apply(a.e, true)
	if a.preStartFunc != nil {
		a.preStartFunc()
	}
	svc := &http.Server{
		Addr:    fmt.Sprintf(":%d", Env.Port),
		Handler: a.e,
	}
	go func() {
		if err := svc.ListenAndServe(); err != nil && !errors.Is(err, http.ErrServerClosed) {
			log.Fatalf("Application start error, %s", err.Error())
		}
	}()
	log.Infof("Application start success on Ports:[%d]", Env.Port)
	quit := make(chan os.Signal)
	signal.Notify(quit, syscall.SIGTERM, syscall.SIGINT)
	<-quit
	log.Info("Shutdown server ...")
	if a.preStopFunc != nil {
		a.preStopFunc()
	}
	ctx, cancelFunc := context.WithTimeout(context.Background(), a.exitDelay)
	defer cancelFunc()
	if err := svc.Shutdown(ctx); err != nil {
		log.Fatalf("Server shutdown failure, %s", err.Error())
	}
	if a.postStopFunc != nil {
		a.postStopFunc()
	}
	log.Info("Server exiting ...")
}

// ReadConfig Read configuration
// v config struct pointer
func (a *App) ReadConfig(v interface{}) *App {
	if err := a.configReader.Unmarshal(v); err != nil {
		log.Fatalf("read config error, %s", err.Error())
	}
	return a
}

// PreApply triggered before mvc starts, Before the project starts.
// This is where you can provide basic services, such as set beans.
// Of course, you can also perform logic here that doesn't require obtaining beans.
func (a *App) PreApply(f func()) *App {
	if f == nil {
		log.Fatalf("apply before func cannot be null.")
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
	a.postStopFunc = f
	return a
}

// ExitDelay Graceful exit time(default 3s), when reached to shut down the server and trigger PostStop().
func (a *App) ExitDelay(time time.Duration) *App {
	a.exitDelay = time
	return a
}
