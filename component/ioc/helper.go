package ioc

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/archine/gin-plus/v4/component/ioc/bean"
	"github.com/archine/gin-plus/v4/internal/container"
	"github.com/archine/gin-plus/v4/internal/vars/sysctr"
)

// RegisterBeanDef registers a candidate bean, typically from an init function.
// The bean is instantiated by the caller but becomes effective only when the container
// refreshes: its Condition is evaluated, then `value` / `autowire` fields are injected
// and BeanPostConstruct is invoked after those of its dependencies.
//
// If instance implements bean.AbstractBean, its BeanName and Condition are honored;
// an empty name falls back to package.StructName.
//
// Panics if instance is not a non-nil struct pointer, if the name is already taken,
// or if the container has already been refreshed.
func RegisterBeanDef(instance any) {
	def, err := container.NewBeanDef("", instance)
	if err != nil {
		panic("[BeanRegistry] registration failed: " + err.Error())
	}

	if ab, ok := instance.(bean.AbstractBean); ok {
		if name := ab.BeanName(); name != "" {
			def.Name = name
		}
		def.Condition = ab.Condition
	}

	sysctr.BeanRegistry.Register(def)
}

// RegisterBean registers an already initialized instance as a singleton bean, effective immediately.
// The instance is neither injected nor post-constructed. It is resolvable by name, by its struct
// type and by every interface it implements. An empty name falls back to package.StructName.
//
// Panics if instance is not a non-nil struct pointer or if the name is already taken.
func RegisterBean(name string, instance any) {
	sysctr.Container.RegisterBean(name, instance)
}

// GetBean retrieves a bean by name. It returns false if the bean does not exist
// or is not of type T.
func GetBean[T any](name string) (T, bool) {
	raw, exist := sysctr.Container.GetBean(name)
	if !exist {
		var zero T
		return zero, false
	}

	val, ok := raw.(T)
	return val, ok
}

// GetBeanByType retrieves the single bean assignable to T (a struct pointer or an interface).
// It returns false if no bean matches.
//
// Panics if several beans match T: the lookup is ambiguous, which is a programming error.
// Use GetBean with an explicit name or GetAllBeansByType instead.
func GetBeanByType[T any]() (T, bool) {
	var zero T
	raw, err := sysctr.Container.GetBeanByType(reflect.TypeFor[T]())
	if err != nil {
		if errors.Is(err, container.ErrAmbiguousBean) {
			panic(fmt.Sprintf("[IOC] %v", err))
		}
		return zero, false
	}

	val, ok := raw.(T)
	return val, ok
}

// GetAllBeansByType retrieves all beans assignable to T in registration order.
// It returns nil if no bean matches.
func GetAllBeansByType[T any]() []T {
	rawList := sysctr.Container.GetAllBeansByType(reflect.TypeFor[T]())
	if len(rawList) == 0 {
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
