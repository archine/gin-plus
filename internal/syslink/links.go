package syslink

import (
	_ "unsafe"

	"github.com/archine/gin-plus/v4/component/ioc"
	"github.com/archine/gin-plus/v4/component/gplog"
)

//go:linkname SetGlobalLogger github.com/archine/gin-plus/v4/component/gplog.setLogger
func SetGlobalLogger(l gplog.Logger)

//go:linkname RefreshContainer github.com/archine/gin-plus/v4/component/ioc.refresh
func RefreshContainer()

//go:linkname GetContainer github.com/archine/gin-plus/v4/component/ioc.getContainer
func GetContainer() *ioc.Container
