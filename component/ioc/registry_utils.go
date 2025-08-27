package ioc

import (
	"reflect"

	"github.com/archine/gin-plus/v4/component/ioc/bean"

	"github.com/archine/gin-plus/v4/internal/container"
	"github.com/archine/gin-plus/v4/internal/vars/sysctr"
)

// RegisterBeanDef registers a bean definition in the IoC container registry.
// It accepts a struct pointer as input and extracts its type information to create a BeanDef.
// If the struct implements the AbstractBean interface, it uses the provided bean name and prototype status.
// If the input is nil or not a struct pointer, it panics.
//
// Parameters:
//   - instance: a struct pointer to be registered.
func RegisterBeanDef(instance any) {
	if instance == nil {
		return
	}

	typ := reflect.TypeOf(instance)
	if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Struct {
		panic("bean instance must be a pointer to a struct")
	}

	def := &container.BeanDef{
		Type:       typ,
		Value:      instance,
		OriginType: typ.Elem(),
	}

	var beanName string
	if ib, ok := instance.(bean.AbstractBean); ok {
		beanName, def.IsPrototype = ib.BeanName(), ib.IsPrototype()
	}

	sysctr.Container.RegisterBeanDef(beanName, def)
}
