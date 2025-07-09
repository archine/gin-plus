package container

import (
	"fmt"
	"github.com/archine/gin-plus/v4/component/ioc"
	"github.com/archine/gin-plus/v4/internal/container/definition"
	"reflect"
	"sync"

	"github.com/archine/gin-plus/v4/component/gplog"
)

var (
	once sync.Once
)

// Refresh refreshes the container by processing all bean definitions.
// It creates beans in the order defined by their dependencies and clears the cache after creation.
// This method should be called only once, typically during application startup.
// It ensures that all beans are created and dependencies are injected correctly.
// If any bean creation fails, it logs a fatal error and stops the application.
func (c *Container) Refresh() {
	once.Do(func() {
		if len(definition.Cache) == 0 {
			return
		}
		createOrderBeanNames, err := c.processDefinition()
		if err != nil {
			gplog.Fatal("Failed to process bean definitions: " + err.Error())
		}

		for _, beanName := range createOrderBeanNames {
			if err = c.createBean(beanName); err != nil {
				gplog.Fatal(fmt.Sprintf("Failed to create bean '%s': %s", beanName, err.Error()))
			}
		}

		definition.Cache = nil // clear definitions to avoid memory leaks
		beanType = nil         // clear bean type to avoid memory leaks
	})
}

// createBean creates a bean instance and performs dependency injection
func (c *Container) createBean(beanName string) error {
	_, found := c.beans[beanName]
	if found {
		// bean already created
		return nil
	}

	def, found := definition.Cache[beanName]
	if !found {
		return fmt.Errorf("bean definition not found for '%s'", beanName)
	}

	if err := c.inject(reflect.ValueOf(def.Bean).Elem(), def.AutowireFields); err != nil {
		return fmt.Errorf("failed to inject dependencies for bean '%s': %w", beanName, err)
	}

	// call post-construct if available
	if postConstruct, ok := def.Bean.(ioc.BeanPostConstruct); ok {
		postConstruct.BeanPostConstruct(beanName)
	}

	c.beans[beanName] = &BeanDef{
		typ:            def.Type,
		value:          def.Bean,
		isPrototype:    def.IsPrototype,
		autowireFields: def.AutowireFields[:len(def.AutowireFields)],
	}
	return nil
}

// createPrototypeBean creates a prototype bean instance
func (c *Container) createPrototypeBean(def *BeanDef) (any, error) {
	newBeanValueOf := reflect.New(def.typ)

	err := c.inject(newBeanValueOf.Elem(), def.autowireFields)
	if err != nil {
		return nil, fmt.Errorf("failed to inject dependencies for bean '%s': %w", def.typ.Name(), err)
	}

	if postConstruct, ok := newBeanValueOf.Interface().(ioc.BeanPostConstruct); ok {
		postConstruct.BeanPostConstruct(def.typ.Name())
	}

	return newBeanValueOf.Interface(), err
}

// injectDependency injects a single dependency into a bean field
func (c *Container) inject(beanValueOf reflect.Value, fields []*definition.AutowireField) error {
	for _, field := range fields {
		var bean any

		if field.AutowireTag == "-" {
			// look for field type in the container
			foundBean, err := c.GetBeanByType(field.Field.Type)
			if err != nil {
				return err
			}
			bean = foundBean
		} else {
			foundBean, exist := c.GetBean(field.AutowireTag)
			if !exist {
				return fmt.Errorf("bean '%s' not found for field '%s'", field.AutowireTag, field.Field.Name)
			}
			bean = foundBean
		}

		fieldValueOf := beanValueOf.Field(field.Index)

		if field.IsInterface {
			if !reflect.TypeOf(bean).Implements(field.Field.Type) {
				return fmt.Errorf("bean '%s' does not implement interface '%s'",
					field.AutowireTag, field.Field.Type.String())
			}
			fieldValueOf.Set(reflect.ValueOf(bean))
			continue
		}

		if !reflect.TypeOf(bean).AssignableTo(field.Field.Type) {
			return fmt.Errorf("type mismatch: expected '%s', got '%s'",
				field.Field.Type.String(), reflect.TypeOf(bean).String())
		}

		fieldValueOf.Set(reflect.ValueOf(bean))
	}

	return nil
}
