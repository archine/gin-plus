package ioc

import (
	"fmt"
	"github.com/archine/gin-plus/v4/internal/container/registry"
	"github.com/archine/gin-plus/v4/util/strutil"
	"reflect"
)

// PreRegisterBean pre-registers a bean definition in the IoC container for later instantiation.
// This function is designed for struct pointers that embed either Bean or mvc.Controller,
// allowing them to be managed by the IoC container.
//
// Args:
//   - instance: a struct pointer that embeds Bean or mvc.Controller
//
// Usage Notes:
//   - This method only registers the bean definition; actual bean instances are created during container refresh.
//   - Typically, only root beans (such as controllers) need manual pre-registration.
//     The container will automatically resolve and instantiate all dependent beans during root bean creation,
//     provided those dependencies also embed Bean or mvc.Controller.
//   - While you can pre-register all beans to avoid reflection overhead during instantiation,
//     this approach increases code complexity and is generally unnecessary.
//   - For root beans with interface-type dependencies, ensure all interface implementations
//     are pre-registered beforehand, otherwise dependency injection will fail.
//
// Returns:
//   - error: nil on success, or an error if registration fails
func PreRegisterBean(instance any) error {
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
