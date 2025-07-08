package app

import (
	"github.com/archine/gin-plus/v4/component/config"
	"github.com/archine/gin-plus/v4/internal/container"
	"reflect"
)

type Context struct {
	container *container.Container
	cf        config.Configure
}

// NewContext creates a new application context with an IoC container and configuration provider.
func NewContext(ct *container.Container) *Context {
	return &Context{
		container: ct,
	}
}

// SetConfigure sets the configuration provider for the application context.
func (c *Context) SetConfigure(cf config.Configure) {
	if cf == nil {
		panic("configuration provider cannot be nil")
	}
	c.cf = cf
}

// GetConfigure returns the current configuration of the application context.
func (c *Context) GetConfigure() config.Configure {
	if c.cf == nil {
		panic("configuration provider is nil, please use WithConfigure() to set a configuration provider")
	}
	return c.cf
}

// GetBean retrieves a bean from the IoC container by its name.
func (c *Context) GetBean(name string) (any, bool) {
	if c.container == nil {
		panic("container is nil, please ensure the IoC container is initialized")
	}
	return c.container.GetBean(name)
}

// GetBeanByType retrieves a bean from the IoC container by its type.
func (c *Context) GetBeanByType(typ reflect.Type) (any, error) {
	if c.container == nil {
		panic("container is nil, please ensure the IoC container is initialized")
	}
	return c.container.GetBeanByType(typ)
}

// RegisterBean manually registers a bean instance to the IOC container.
// This method allows registration of pre-created objects that don't need to go through
// the automatic bean creation process.
//
// Args:
//   - name: bean name for registration. If empty, defaults to the struct name with first letter lowercase
//   - objPtr: pointer to the struct instance to register (must not be nil)
//   - implementedTypes: optional interface types that the object implements for type-based lookup
func (c *Context) RegisterBean(name string, objPtr any, implementedTypes ...reflect.Type) error {
	if c.container == nil {
		panic("container is nil, please ensure the IoC container is initialized")
	}
	return c.container.RegisterBean(name, objPtr, implementedTypes...)
}
