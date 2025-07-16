package sysconf

import "github.com/archine/gin-plus/v4/component/gpconf"

var (
	ProjectConfigure gpconf.Configure // ProjectConfigure is the global configuration for the project, initialized in main.go
)

func InitConfigure(configureFunc func() gpconf.Configure) {
	if configureFunc == nil {
		configureFunc = gpconf.NewLocalFileConfigure
	}

	ProjectConfigure = configureFunc()
}
