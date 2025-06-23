package ioc

import (
	"errors"
	"fmt"
	"github.com/archine/gin-plus/v4/internal/container"
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

	return container.SetBean(beanName, bean)
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

// GetBeanByType retrieves a bean from the container by its type.
func GetBeanByType(beanStruct any) any {
	if beanStruct == nil {
		return nil
	}

	return container.GetBeanByType(beanStruct)
}

// GetBeanByName retrieves a bean from the container by its registered name
// Returns nil if no bean is found with the given name
func GetBeanByName(beanName string) any {
	if beanName == "" {
		return nil
	}

	return container.GetBeanByName(beanName)
}
