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
