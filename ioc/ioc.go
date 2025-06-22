package ioc

import (
	"errors"
	"fmt"
	"github.com/archine/gin-plus/v4/internal/container"
	"reflect"
)

// Bean provides a default implementation of FactoryBean interface
// This should be embedded in other structs, and those structs should override CreateBean method
type Bean struct{}

func (b *Bean) BeanPostConstruct() {}

// SetBean manually registers a bean in the IoC container
// This is mainly used for runtime registration of beans that don't implement ioc.Bean
func SetBean(beanName string, bean any) error {
	if bean == nil {
		return errors.New("failed to set bean, instance cannot be nil")
	}

	beanTyp := reflect.TypeOf(bean)
	if beanTyp.Kind() != reflect.Ptr {
		return fmt.Errorf("bean must be a non-nil pointer, got %s", beanTyp.Kind().String())
	}
	beanTyp = beanTyp.Elem()

	if beanName == "" {
		beanName = beanTyp.Name()
	}

	container.Mutex.Lock()
	defer container.Mutex.Unlock()

	if _, exists := container.BeanTypeName[beanTyp]; exists {
		return fmt.Errorf("bean type '%s' is already registered", beanTyp.String())
	}
	// Check if the bean is already registered
	if _, exists := container.BeanCache[beanName]; exists {
		return fmt.Errorf("bean name '%s' is already exists", beanName)
	}

	container.BeanCache[beanName] = bean
	container.BeanTypeName[beanTyp] = beanName

	return nil
}

// SetBeans registers multiple beans in the IoC container.
func SetBeans(beans ...any) error {
	if len(beans) == 0 {
		return nil
	}

	for _, bean := range beans {
		if err := SetBean("", bean); err != nil {
			return fmt.Errorf("failed to set bean: %w", err)
		}
	}

	return nil
}

// GetBean retrieves a bean from the container by its type.
func GetBean(beanStruct any) any {
	if beanStruct == nil {
		return nil
	}

	typ := reflect.TypeOf(beanStruct)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	container.Mutex.RLock()
	defer container.Mutex.RUnlock()

	if beanName, exists := container.BeanTypeName[typ]; exists {
		return container.BeanCache[beanName]
	}

	return nil
}

// GetBeanByName retrieves a bean from the container by its registered name
// Returns nil if no bean is found with the given name
func GetBeanByName(beanName string) any {
	if beanName == "" {
		return nil
	}

	container.Mutex.RLock()
	defer container.Mutex.RUnlock()

	return container.BeanCache[beanName]
}
