package syslink

import (
	"github.com/gin-gonic/gin"
	_ "unsafe"

	"github.com/archine/gin-plus/v4/component/gplog"
	"github.com/archine/gin-plus/v4/component/ioc"
)

//go:linkname SetGlobalLogger github.com/archine/gin-plus/v4/component/gplog.setLogger
func SetGlobalLogger(l gplog.Logger)

//go:linkname RefreshContainer github.com/archine/gin-plus/v4/component/ioc.refresh
func RefreshContainer()

//go:linkname GetContainer github.com/archine/gin-plus/v4/component/ioc.getContainer
func GetContainer() *ioc.Container

//go:linkname ApplyRoute github.com/archine/gin-plus/v4/component/mvc/router.apply
func ApplyRoute(engine *gin.Engine, contextPath string, enableHealth bool) error
