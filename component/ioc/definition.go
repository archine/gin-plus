package ioc

import (
	"fmt"
	"github.com/archine/gin-plus/v4/internal/container/definition"
	"github.com/archine/gin-plus/v4/util/strutil"
	"reflect"
)

// RegisterBeanDefinition registers a bean definition in the IoC container.
// This function is used to register bean definitions that will be managed by the IoC container.
//
// Args:
//   - instances: one or more struct pointers that implement the Bean interface.
func RegisterBeanDefinition(instances ...any) error {
	if len(instances) == 0 {
		return nil
	}

	for _, instance := range instances {
		if instance == nil {
			continue
		}

		ib, ok := instance.(Bean)
		if !ok {
			continue
		}

		typ := reflect.TypeOf(instance)
		if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Struct {
			return fmt.Errorf("type '%s' must be a pointer to a struct", typ.Name())
		}

		typ = typ.Elem()

		beanName := ib.BeanName()
		if beanName == "" {
			beanName = strutil.FirstToLower(typ.Name())
		}

		if _, found := definition.Cache[beanName]; found {
			return fmt.Errorf("bean '%s' already registered", beanName)
		}

		def := &definition.BeanDefinition{
			Name:           beanName,
			Type:           typ,
			IsPrototype:    ib.IsPrototype(),
			AutowireFields: make([]*definition.AutowireField, 0, 4),
			Bean:           instance,
		}

		definition.Cache[beanName] = def
	}

	return nil
}
