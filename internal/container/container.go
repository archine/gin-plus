package container

import (
	"fmt"
	"github.com/archine/gin-plus/v4/internal/util"
	"reflect"
	"sync"
	"unsafe"
)

var (
	Mutex           = sync.RWMutex{}
	BeanCache       = make(map[string]any)    // Complete IoC container for managing beans
	BeanTypeName    = make(map[string]string) // Cache for bean types, used to avoid duplicate registrations
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
	var fieldIsIface bool
	var fieldOriginType reflect.Type
	var fieldTypName string

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
		fieldTypName = util.GetTypeKey(fieldOriginType)

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

		var bean any
		if autowire == "-" {
			bean = GetBeanByTypeName(fieldTypName)
		} else {
			bean = GetBeanByName(autowire)
		}

		if bean == nil {
			if autowire != "-" || field.Type.Implements(factoryBeanType) {
				return fmt.Errorf("bean '%s' not found in container and field %s.%s", autowire, structType.Name(), field.Name)
			}
			beanValue, err := createBean(fieldOriginType, fieldTypName)
			if err != nil {
				return fmt.Errorf("failed to create bean via factory for field %s.%s: %w", structType.Name(), field.Name, err)
			}
			fieldVal.Set(beanValue)
		} else {
			beanValue := reflect.ValueOf(bean)
			if fieldIsIface && beanValue.Type().Implements(field.Type) {
				fieldVal.Set(beanValue)
			} else if beanValue.Type().AssignableTo(field.Type) {
				fieldVal.Set(beanValue)
			} else {
				return fmt.Errorf("bean type mismatch for field %s.%s: expected %s, got %s", structType.Name(), field.Name, field.Type.String(), beanValue.Type().String())
			}
		}
	}

	if fb, ok := objPtr.(FactoryBean); ok {
		fb.BeanPostConstruct()
	}

	return nil
}

// SetBean registers a bean in the IoC container.
func SetBean(beanName string, bean any) error {
	beanTyp := reflect.TypeOf(bean)
	if beanTyp.Kind() != reflect.Ptr {
		return fmt.Errorf("bean must be a non-nil pointer, got %s", beanTyp.Kind().String())
	}
	beanTyp = beanTyp.Elem()

	if beanName == "" {
		beanName = beanTyp.Name()
	}

	Mutex.Lock()
	defer Mutex.Unlock()

	// Check if the bean is already registered
	if _, exists := BeanCache[beanName]; exists {
		return fmt.Errorf("bean name '%s' is already exists", beanName)
	}

	BeanCache[beanName] = bean
	BeanTypeName[util.GetTypeKey(beanTyp)] = beanName

	return nil
}

// GetBeanByType retrieves a bean from the container by its type.
func GetBeanByType(bean any) any {
	typ := reflect.TypeOf(bean)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	Mutex.RLock()
	defer Mutex.RUnlock()

	if beanName, exists := BeanTypeName[util.GetTypeKey(typ)]; exists {
		return BeanCache[beanName]
	}

	return nil
}

// GetBeanByTypeName retrieves a bean from the container by its type name.
func GetBeanByTypeName(beanTypeName string) any {
	Mutex.RLock()
	defer Mutex.RUnlock()

	beanName, exists := BeanTypeName[beanTypeName]
	if !exists {
		return nil
	}

	return BeanCache[beanName]
}

// GetBeanByName retrieves a bean from the container by its name.
func GetBeanByName(beanName string) any {
	Mutex.RLock()
	defer Mutex.RUnlock()

	bean, exists := BeanCache[beanName]
	if !exists {
		return nil
	}

	return bean
}

// createBean creates a new bean instance using FactoryBean and injects its dependencies
func createBean(fieldOriginType reflect.Type, typName string) (reflect.Value, error) {
	Mutex.Lock()

	if beanName, ok := BeanTypeName[typName]; ok {
		if bean, exists := BeanCache[beanName]; exists {
			return reflect.ValueOf(bean), nil
		}
		Mutex.Unlock()
	}

	beanName := factoryBeanType.Name()
	newBean := reflect.New(fieldOriginType)

	BeanCache[beanName] = newBean.Interface()
	BeanTypeName[typName] = beanName
	Mutex.Unlock()

	// Recursively inject dependencies into the newly created bean
	if err := Inject(newBean); err != nil {
		Mutex.Lock()
		delete(BeanCache, beanName)
		delete(BeanTypeName, typName)
		Mutex.Unlock()
		return reflect.ValueOf(nil), fmt.Errorf("failed to inject dependencies for bean '%s': %w", beanName, err)
	}

	return newBean, nil
}
