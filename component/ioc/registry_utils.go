package ioc

import (
	"fmt"
	"github.com/archine/gin-plus/v4/internal/container/registry"
	"github.com/archine/gin-plus/v4/util/strutil"
	"reflect"
)

// RegisterBeanDefinition registers a bean definition in the IoC container.
// This function is used to register struct pointers that implement the AbstractBean interface for management by the IoC container.
//
// Args:
//   - instance: a struct pointer that implements the AbstractBean interface.
//
// Notes:
//   - In most cases, you only need to manually register root bean instances (such as controller beans).
//     When the container creates a root bean, it will automatically resolve and instantiate all required dependent beans,
//     as long as these dependencies implement the AbstractBean interface.
//   - You may also choose to manually register all beans to avoid the performance overhead of reflection-based instantiation,
//     but this increases code complexity and is generally unnecessary.
//   - If a root bean contains fields of interface type that require dependency injection,
//     please ensure all implementations of these interfaces are registered in advance using this method before
func RegisterBeanDefinition(instance any) error {
	if instance == nil {
		return nil
	}

	ib, ok := instance.(AbstractBean)
	if !ok {
		return fmt.Errorf("instance '%s' does not implement AbstractBean interface", reflect.TypeOf(instance).Name())
	}

	typ := reflect.TypeOf(instance)
	if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Struct {
		return fmt.Errorf("type '%s' must be a pointer to a struct", typ.Name())
	}

	originTyp := typ.Elem()

	if registry.LookupType(originTyp) {
		return fmt.Errorf("type '%s' has already been registered", typ.Name())
	}

	beanName := ib.BeanName()
	if beanName == "" {
		beanName = strutil.FirstToLower(originTyp.Name())
	}

	registry.RegisterBeanDefinition(&registry.BeanDefinition{
		PtrType:     typ,
		Name:        beanName,
		Bean:        instance,
		OriginType:  originTyp,
		IsPrototype: ib.IsPrototype(),
	})

	return nil
}
