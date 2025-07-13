package app

import (
	"github.com/archine/gin-plus/v4/component/config"
	"github.com/archine/gin-plus/v4/internal/container"
	"reflect"
)

type Context struct {
	cf        config.Configure
	container *container.Container
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
func (c *Context) GetBeanByType(typ reflect.Type) (any, bool) {
	if c.container == nil {
		panic("container is nil, please ensure the IoC container is initialized")
	}
	return c.container.GetBeanByType(typ)
}

// GetAllBeansByType retrieves all beans from the IoC container that match a specific type.
func (c *Context) GetAllBeansByType(typ reflect.Type) ([]any, bool) {
	if c.container == nil {
		panic("container is nil, please ensure the IoC container is initialized")
	}
	return c.container.GetAllBeansByType(typ)
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
	if c.container == nil {
		panic("container is nil, please ensure the IoC container is initialized")
	}
	return c.container.RegisterBean(name, instance, itypes...)
}
