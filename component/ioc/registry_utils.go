package ioc

import (
	"github.com/archine/gin-plus/v4/component/ioc/bean"
	"reflect"

	"github.com/archine/gin-plus/v4/internal/container"
	"github.com/archine/gin-plus/v4/internal/vars/sysctr"
)

// BeanDefinition is an alias for container.BeanDef. Aliasing this type dramatically
// improves the navigability of this package's API documentation.
type BeanDefinition = container.BeanDef

// AutowireField is an alias for container.AutowireField. This alias is used to
// improve the readability of the API documentation for this package.
type AutowireField = container.AutowireField

// RegisterBeanDef registers a bean definition in the IoC container registry
//
// Parameters:
//   - instance: a struct pointer to be registered.
//
// Usage Notes:
//   - Typically, only top-level objects require manual registration—these are objects that are not used as field types depended on by other structures.
//   - When the container parses the Bean definition, it scans the fields with injection tags in the structure;
//     If the type of dependent field is not registered in the registry, the container will automatically generate and register a Bean definition for it.
//   - Although all objects can be manually registered to reduce the reflection overhead during instantiation,
//     this approach is usually unnecessary and will increase the complexity of the code.
func RegisterBeanDef(instance any) {
	if instance == nil {
		return
	}

	typ := reflect.TypeOf(instance)
	if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Struct {
		panic("bean instance must be a pointer to a struct")
	}

	originTyp := typ.Elem()

	if sysctr.Container.LookupType(originTyp) {
		return
	}

	def := &BeanDefinition{
		Type:       typ,
		Value:      instance,
		OriginType: originTyp,
	}

	var beanName string
	if ib, ok := instance.(bean.AbstractBean); ok {
		beanName, def.IsPrototype = ib.BeanName(), ib.IsPrototype()
	}

	sysctr.Container.RegisterBeanDef(beanName, def)
}
