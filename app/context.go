package app

import (
	"reflect"

	"github.com/archine/gin-plus/v4/component/config"
)

// ApplicationContext the interface defines methods for managing the application context,
type ApplicationContext interface {
	// GetConfigProvider gets the configuration provider for the application context.
	GetConfigProvider() config.Provider

	// GetBean retrieves a bean from the IoC container by its name.
	// It returns the bean instance and a boolean indicating if the bean was found.
	GetBean(name string) (any, bool)

	// GetBeanByType retrieves a bean from the IoC container by its type.
	// It returns the bean instance and a boolean indicating if the bean was found.
	GetBeanByType(typ reflect.Type) (any, bool)

	// GetAllBeansByType retrieves all beans from the IoC container that match a specific type.
	// It returns a slice of beans and a boolean indicating if any beans were found.
	GetAllBeansByType(typ reflect.Type) ([]any, bool)

	// RegisterBean registers an already instantiated bean instance into the IOC container.
	// Parameters:
	//   - name: the bean name for registration.
	//   - instance: a pointer to the struct instance that has already been instantiated.
	//   - itypes: optional interface types that the object implements, used for type-based lookup.
	// Note:
	//   - According to the IOC container design, all beans registered by this method are singletons.
	//   - If a bean with the specified name already exists, an error will be returned.
	//   - During registration, type-to-bean-name mappings are automatically established to support type-based bean retrieval.
	RegisterBean(name string, instance any, itypes ...reflect.Type)
}
