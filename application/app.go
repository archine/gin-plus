package application

import (
	"fmt"
	"github.com/archine/gin-plus/v2/exception"
	"github.com/archine/gin-plus/v2/mvc"
	"github.com/archine/gin-plus/v2/plugin"
	"github.com/archine/ioc"
	"github.com/gin-contrib/cors"
	"github.com/gin-gonic/gin"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
	"net/http"
)

// App application instance
type App struct {
	e           *gin.Engine
	applyBefore func()
	startBefore func()
}

// New a clean project
func New() *App {
	LoadApplicationConfigFile()
	plugin.InitLog(Env.LogLevel)
	if Env.LogLevel == "debug" {
		gin.SetMode(gin.DebugMode)
	} else {
		gin.SetMode(gin.ReleaseMode)
	}
	engine := gin.New()
	engine.MaxMultipartMemory = Env.MaxFileSize
	engine.RemoveExtraSlash = true
	ioc.SetBeans(engine)
	return &App{e: engine}
}

// Default create a default project that integrates logging, exception interception, and cross-domain by default
func Default() *App {
	LoadApplicationConfigFile()
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
		AllowMethods:     []string{"PUT", "PATCH", "POST", "GET", "DELETE", "OPTIONS"},
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
		e: engine,
	}
}

// Run the main program entry
// interceptors: link{mvc.MethodInterceptor}
func (a *App) Run(interceptors ...mvc.MethodInterceptor) {
	if a.applyBefore != nil {
		a.applyBefore()
	}
	if len(interceptors) > 0 {
		a.e.Use(func(context *gin.Context) {
			is := interceptors
			for _, interceptor := range is {
				if interceptor.Predicate(context.Request) {
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
	svc := &http.Server{
		Addr:    fmt.Sprintf(":%d", Env.Port),
		Handler: a.e,
	}
	if a.startBefore != nil {
		a.startBefore()
	}
	log.Infof("Application start success on Ports:[%d]", Env.Port)
	if err := svc.ListenAndServe(); err != nil {
		log.Fatalf("Application start error, %s", err.Error())
	}
}

// ReadConfig Read configuration
// v: config struct pointer
func (a *App) ReadConfig(v interface{}) *App {
	if err := viper.Unmarshal(v); err != nil {
		log.Fatalf("read config error, %s", err.Error())
	}
	return a
}

// ApplyBefore triggered before mvc starts, Before the project starts.
// This is where you can provide basic services, such as set beans.
// Of course, you can also perform logic here that doesn't require obtaining beans.
func (a *App) ApplyBefore(f func()) *App {
	if f == nil {
		log.Fatalf("apply before func cannot be null.")
	}
	a.applyBefore = f
	return a
}

// StartBefore The last event before the project starts, dependency injection is all finished and ready to run.
// You can execute any logic here.
func (a *App) StartBefore(f func()) *App {
	if f == nil {
		log.Fatalf("start after func cannot be null.")
	}
	a.startBefore = f
	return a
}
