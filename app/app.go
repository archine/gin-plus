package app

import (
	"github.com/archine/gin-plus/v4/app/event"
	"github.com/archine/gin-plus/v4/component/config"
	"github.com/archine/gin-plus/v4/component/gplog/iface"
	"github.com/archine/gin-plus/v4/component/mvc/ctrl"
	"time"

	"github.com/gin-gonic/gin"
)

type App struct {
	engine       *gin.Engine
	exitDelay    time.Duration
	middlewares  []gin.HandlerFunc
	interceptors []ctrl.MethodInterceptor
	eventManager *event.Manager
	configure    *config.Configure
	logger       iface.Logger
}
