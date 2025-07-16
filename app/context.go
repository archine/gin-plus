package app

import (
	"github.com/archine/gin-plus/v4/component/gpconf"
	"github.com/archine/gin-plus/v4/internal/vars/sysconf"
	"github.com/archine/gin-plus/v4/internal/vars/syscontainer"
	"reflect"
)

type Context struct{}

// NewContext creates a new application context with an IoC container and configuration provider.
func NewContext() *Context {
	return &Context{}
}

// GetConfigure returns the current configuration of the application context.
func (c *Context) GetConfigure() gpconf.Configure {
	return sysconf.ProjectConfigure
}

// GetBean retrieves a bean from the IoC container by its name.
func (c *Context) GetBean(name string) (any, bool) {
	return syscontainer.Container.GetBean(name)
}

// GetBeanByType retrieves a bean from the IoC container by its type.
func (c *Context) GetBeanByType(typ reflect.Type) (any, bool) {
	return syscontainer.Container.GetBeanByType(typ)
}

// GetAllBeansByType retrieves all beans from the IoC container that match a specific type.
func (c *Context) GetAllBeansByType(typ reflect.Type) ([]any, bool) {
	return syscontainer.Container.GetAllBeansByType(typ)
}

// RegisterBean registers an already instantiated bean instance into the IOC container.
//
// Parameters:
//   - name: the bean name for registration.
//   - instance: a pointer to the struct instance that has already been instantiated.
//   - itypes: optional interface types that the object implements, used for type-based lookup.
//
// Note:
//   - According to the IOC container design, all beans registered by this method are singletons.
//   - If a bean with the specified name already exists, an error will be returned.
//   - During registration, type-to-bean-name mappings are automatically established to support type-based bean retrieval.
func (c *Context) RegisterBean(name string, instance any, itypes ...reflect.Type) error {
	return syscontainer.Container.RegisterBean(name, instance, itypes...)
}
