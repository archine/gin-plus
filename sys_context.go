package gin_plus

import (
	"github.com/archine/gin-plus/v4/app"
	"github.com/archine/gin-plus/v4/component/config"
	"github.com/archine/gin-plus/v4/internal/vars/sysconf"
	"reflect"

	"github.com/archine/gin-plus/v4/internal/vars/sysctr"
)

// sysContext the system context implementation
type sysContext struct{}

func newSysContext() app.ApplicationContext {
	return &sysContext{}
}

func (s *sysContext) GetConfigProvider() config.Provider {
	return sysconf.Provider
}

func (s *sysContext) GetBean(name string) (any, bool) {
	return sysctr.Container.GetBean(name)
}

func (s *sysContext) GetBeanByType(typ reflect.Type) (any, bool) {
	return sysctr.Container.GetBeanByType(typ)
}

func (s *sysContext) GetAllBeansByType(typ reflect.Type) ([]any, bool) {
	return sysctr.Container.GetAllBeansByType(typ)
}

func (s *sysContext) RegisterBean(name string, instance any, itypes ...reflect.Type) {
	sysctr.Container.RegisterBean(name, instance, itypes...)
}
