package ioc

import (
	"fmt"
	"reflect"

	"github.com/archine/gin-plus/v4/component/ioc/bean"
	"github.com/archine/gin-plus/v4/internal/vars/sysconf"

	"github.com/archine/gin-plus/v4/internal/container"
	"github.com/archine/gin-plus/v4/internal/vars/sysctr"
)

// RegisterBeanDef registers a bean definition in the IoC container registry.
// It accepts a struct pointer as input and extracts its type information to create a BeanDef.
// If the struct implements the AbstractBean interface, it uses the provided bean name.
// If no explicit name is provided, the container uses package.structName as the default.
// If the input is nil or not a struct pointer, it panics.
//
// Parameters:
//   - instance: a struct pointer to be registered.
//
// Note:
//   - The function will panic if the input is nil or not a struct pointer.
//   - If a bean with the specified name already exists, an error will be returned.
func RegisterBeanDef(instance any) {
	if instance == nil {
		return
	}

	typ := reflect.TypeOf(instance)
	if typ.Kind() != reflect.Pointer || typ.Elem().Kind() != reflect.Struct {
		panic(fmt.Sprintf("[BeanRegistry] registration failed: '%s' is not a struct pointer", typ.String()))
	}

	var beanName string
	if ib, ok := instance.(bean.AbstractBean); ok {
		if !ib.Condition(sysconf.Provider) {
			return
		}
		beanName = ib.BeanName()
	}

	def := &container.BeanDef{
		Type:       typ,
		Value:      instance,
		OriginType: typ.Elem(),
	}

	sysctr.Container.RegisterBeanDef(beanName, def)
}

// RegisterBean registers an already instantiated bean instance into the IOC container.
// Parameters:
//   - name: the bean name for registration.
//   - instance: a pointer to the struct instance that has already been instantiated.
//   - itypes: optional interface types that the object implements, used for type-based lookup.
//
// Note:
//   - According to the IOC container design, all beans registered by this method are singletons.
//   - If a bean with the specified name already exists, an error will be returned.
//   - During registration, type-to-bean-name mappings are automatically established to support type-based bean retrieval.
func RegisterBean(name string, instance any, itypes ...reflect.Type) {
	sysctr.Container.RegisterBean(name, instance, itypes...)
}

// GetBean retrieves a bean from the IoC container by its name and returns it as a pointer of the specified type.
// If the bean does not exist, it returns nil.
//
// Parameters:
//   - name: the name of the bean to retrieve.
//
// Returns:
//   - a pointer to the bean of type T, or nil if the bean does not exist.
func GetBean[T any](name string) (T, bool) {
	var zero T
	raw, exist := sysctr.Container.GetBean(name)
	if !exist {
		return zero, false
	}

	val, ok := raw.(T)
	return val, ok
}

// GetBeanByType retrieves a bean from the IoC container by its type and returns it as a pointer of the specified type.
// If the bean does not exist, it returns nil.
//
// Returns:
//   - a pointer to the bean of type T, or false if the bean does not exist.
func GetBeanByType[T any]() (T, bool) {
	var zero T
	typ := reflect.TypeFor[T]()

	raw, exist := sysctr.Container.GetBeanByType(typ)
	if !exist {
		return zero, false
	}

	val, ok := raw.(T)
	return val, ok
}

// GetAllBeansByType retrieves all beans from the IoC container that match the specified type and returns them as a slice of pointers of the specified type.
// If no beans match the specified type, it returns nil.
//
// Returns:
//   - a slice of pointers to beans of type T, or nil if no beans match the specified type.
func GetAllBeansByType[T any]() []T {
	typ := reflect.TypeFor[T]()

	rawList, exist := sysctr.Container.GetAllBeansByType(typ)
	if !exist || len(rawList) == 0 {
		return nil
	}

	result := make([]T, 0, len(rawList))
	for _, raw := range rawList {
		if val, ok := raw.(T); ok {
			result = append(result, val)
		}
	}

	return result
}
