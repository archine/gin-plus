package container

import (
	"fmt"
	"github.com/archine/gin-plus/v4/component/gplog"
	"github.com/archine/gin-plus/v4/component/ioc"
	"github.com/archine/gin-plus/v4/component/mvc"
	"github.com/archine/gin-plus/v4/internal/container/definition"
	"reflect"
	"sync"
)

var (
	once     sync.Once
	ctrlType = reflect.TypeOf((*mvc.AbstractController)(nil)).Elem()
	//confRegex = regexp.MustCompile(`\$\{([^:}]+)(?::([^}]*))?}`)
)

// Refresh refreshes the container by processing all bean definitions.
// It creates beans in the order defined by their dependencies and clears the cache after creation.
// This method should be called only once, typically during application startup.
// It ensures that all beans are created and dependencies are injected correctly.
// If any bean creation fails, it logs a fatal error and stops the application.
func (c *Container) Refresh() {
	once.Do(func() {
		err := processDefinitions(c)
		if err != nil {
			gplog.Fatal("Failed to process bean definitions: " + err.Error())
		}

		if err = initBean(c); err != nil {
			gplog.Fatal(err.Error())
		}

		refresh(c)
	})
}

func initBean(c *Container) error {
	for beanName, bean := range c.beans {
		if bean.ready {
			continue
		}

		if err := inject(c, reflect.ValueOf(bean.value).Elem(), bean.autowireFields); err != nil {
			return fmt.Errorf("failed to init bean '%s': %w", beanName, err)
		}
	}

	return nil
}

// createPrototypeBean creates a prototype bean instance
func createPrototypeBean(c *Container, def *BeanDef) any {
	newBeanValueOf := reflect.New(def.originTyp)

	err := inject(c, newBeanValueOf.Elem(), def.autowireFields)
	if err != nil {
		gplog.Fatal(fmt.Sprintf("Failed to create prototype bean '%s': %s", def.originTyp.Name(), err.Error()))
	}

	if postConstruct, ok := newBeanValueOf.Interface().(ioc.BeanPostConstruct); ok {
		postConstruct.BeanPostConstruct()
	}

	return newBeanValueOf.Interface()
}

// injectDependency injects a single dependency into a bean field
func inject(c *Container, beanValueOf reflect.Value, fields []*definition.AutowireField) error {
	for _, field := range fields {
		var bean any
		var exist bool

		if field.AutowireTag == "-" {
			// look for field type in the container
			bean, exist = c.GetBeanByType(field.Field.Type)
		} else {
			bean, exist = c.GetBean(field.AutowireTag)
		}
		if !exist {
			return fmt.Errorf("not found bean for field '%s' in '%s'", field.Name, beanValueOf.Type().String())
		}

		if field.IsInterface {
			if !reflect.TypeOf(bean).Implements(field.Field.Type) {
				return fmt.Errorf("bean '%s' does not implement field '%s' (type: %s)",
					field.AutowireTag, field.Name, field.Field.Type.String())
			}
			setFieldValue(beanValueOf.Field(field.Index), field.Field, bean)
			continue
		}

		if !reflect.TypeOf(bean).AssignableTo(field.Field.Type) {
			return fmt.Errorf("bean '%s' (type: %s) cannot be assigned to field '%s' (type: %s)",
				field.AutowireTag, reflect.TypeOf(bean).String(), field.Name, field.Field.Type.String())
		}

		setFieldValue(beanValueOf.Field(field.Index), field.Field, bean)
	}

	return nil
}

func setFieldValue(fieldValue reflect.Value, field reflect.StructField, value any) {
	if field.IsExported() {
		fieldValue.Set(reflect.ValueOf(value))
	} else {
		// Use unsafe to set unexported field value
		elem := reflect.NewAt(field.Type, fieldValue.Addr().UnsafePointer()).Elem()
		elem.Set(reflect.ValueOf(value))
	}
}

func refresh(c *Container) {
	for beanName, bean := range c.beans {
		if bean.ready {
			continue
		}

		if postConstruct, ok := bean.value.(ioc.BeanPostConstruct); ok {
			postConstruct.BeanPostConstruct()
		}

		if _, ok := bean.value.(mvc.AbstractController); ok {
			c.typeMapping[ctrlType] = append(c.typeMapping[ctrlType], beanName)
		}

		bean.ready = true
	}
}
