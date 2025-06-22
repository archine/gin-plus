package container

import (
	"fmt"
	"reflect"
	"sync"
	"unsafe"
)

var (
	Mutex           = sync.RWMutex{}
	BeanCache       = make(map[string]any)          // Complete IoC container for managing beans
	BeanTypeName    = make(map[reflect.Type]string) // Cache for bean types, used to avoid duplicate registrations
	factoryBeanType = reflect.TypeOf((*FactoryBean)(nil)).Elem()
)

// FactoryBean defines the interface for beans that can be managed by the IoC container
// Beans implementing this interface can create instances of themselves and provide naming
type FactoryBean interface {
	// BeanPostConstruct is called after the bean is created and dependencies are injected
	BeanPostConstruct()
}

// Inject injects all dependency fields of the given object
// The object must be a non-nil pointer to a struct
func Inject(objPtr any) error {
	val := reflect.ValueOf(objPtr)
	if val.Kind() != reflect.Ptr || val.IsNil() {
		return fmt.Errorf("injection target must be a non-nil pointer, got %T", objPtr)
	}

	structElem := val.Elem()
	if structElem.Kind() != reflect.Struct {
		return fmt.Errorf("injection target must be a pointer to struct, got pointer to %s", structElem.Kind().String())
	}
	structType := structElem.Type()

	var fieldKind reflect.Kind
	var fieldOriginType reflect.Type
	var fieldIsIface bool

	for i := 0; i < structType.NumField(); i++ {
		field := structType.Field(i)

		// Skip anonymous/embedded fields as they are difficult to handle properly
		if field.Anonymous {
			continue
		}

		// Get autowire tag value, skip if not present.
		// This tag indicates that the field should be injected with a bean from the container.
		// The tag value is the name of the bean to be searched for. If it is "-", it indicates lookup by type.
		// When the field type is an interface, the tag value must be explicitly set.
		//
		// Examples: `autowire:"myBean"` or `autowire:"-"`.
		autowire := field.Tag.Get("autowire")
		if autowire == "" {
			continue
		}

		fieldKind = field.Type.Kind()
		fieldIsIface = fieldKind == reflect.Interface
		fieldOriginType = field.Type.Elem()

		// Only inject interface types or pointer to struct types
		// Skip primitive types, slices, maps, etc.
		if !fieldIsIface && !(fieldKind == reflect.Ptr && fieldOriginType.Kind() == reflect.Struct) {
			continue
		}

		if autowire == "-" {
			if fieldIsIface {
				return fmt.Errorf("field %s.%s is an interface and requires an explicit bean name", structType.Name(), field.Name)
			}
		}

		fieldVal := structElem.Field(i)
		if !fieldVal.CanSet() {
			// If the field is unexported, we need to use unsafe pointer to set it
			if !fieldVal.CanAddr() {
				return fmt.Errorf("field %s.%s is unexported and cannot be addressed for injection", structType.Name(), field.Name)
			}
			fieldVal = reflect.NewAt(field.Type, unsafe.Pointer(fieldVal.UnsafeAddr())).Elem()
		}

		Mutex.RLock()
		bean, exists := BeanCache[autowire]
		Mutex.RUnlock()

		if exists {
			beanVal := reflect.ValueOf(bean)
			if fieldIsIface {
				// Check if bean implements the required interface
				if beanVal.Type().Implements(field.Type) {
					fieldVal.Set(beanVal)
				} else {
					return fmt.Errorf("bean '%s' does not implement required interface %s for field %s.%s",
						autowire, field.Type.String(), structType.Name(), field.Name)
				}
			} else if beanVal.Type().AssignableTo(field.Type) {
				fieldVal.Set(beanVal)
			} else {
				return fmt.Errorf("bean '%s' cannot be assigned to field %s.%s of type %s",
					autowire, structType.Name(), field.Name, field.Type.String())
			}

		} else {
			if autowire != "-" {
				return fmt.Errorf("bean '%s' not found in container and field %s.%s", autowire, structType.Name(), field.Name)
			}

			if field.Type.Implements(factoryBeanType) {
				created, err := createBean(fieldOriginType)
				if err != nil {
					return fmt.Errorf("failed to create bean via factory for field %s.%s: %w", structType.Name(), field.Name, err)
				}
				fieldVal.Set(reflect.ValueOf(created))
				continue
			}
			return fmt.Errorf("bean '%s' not found in container and field %s.%s does not implement ioc.Bean",
				autowire, structType.Name(), field.Name)
		}
	}

	if fb, ok := objPtr.(FactoryBean); ok {
		fb.BeanPostConstruct()
	}

	return nil
}

// createBean creates a new bean instance using FactoryBean and injects its dependencies
func createBean(fieldOriginType reflect.Type) (any, error) {
	Mutex.Lock()

	if beanName, ok := BeanTypeName[fieldOriginType]; ok {
		Mutex.Unlock()
		return BeanCache[beanName], nil
	}

	beanName := factoryBeanType.Name()
	newBean := reflect.New(fieldOriginType).Interface()

	BeanCache[beanName] = newBean
	BeanTypeName[fieldOriginType] = beanName
	Mutex.Unlock()

	// Recursively inject dependencies into the newly created bean
	if err := Inject(newBean); err != nil {
		Mutex.Lock()
		delete(BeanCache, beanName)
		delete(BeanTypeName, fieldOriginType)
		Mutex.Unlock()
		return nil, fmt.Errorf("failed to inject dependencies for bean '%s': %w", beanName, err)
	}

	return newBean, nil
}
