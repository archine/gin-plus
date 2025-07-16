package syslog

import (
	"github.com/archine/gin-plus/v4/component/gpconf"
	"github.com/archine/gin-plus/v4/component/gplog/gplogcore"
	"github.com/archine/gin-plus/v4/internal/vars/sysconf"
)

var (
	GlobalLog gplogcore.Logger // GlobalLog is the global logger instance used throughout the application.
)

func InitializeLogger(loggerFunc func(conf gpconf.Configure) gplogcore.Logger) {
	if loggerFunc == nil {
		loggerFunc = gplogcore.NewZapLogger
	}

	GlobalLog = loggerFunc(sysconf.ProjectConfigure)
}
