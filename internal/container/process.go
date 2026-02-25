package container

import (
	"fmt"
	"github.com/archine/gin-plus/v4/component/ioc/bean"
	"github.com/archine/gin-plus/v4/component/mvc"
	"github.com/archine/gin-plus/v4/internal/container/injector"
	"reflect"
)

// analyzeDefinition analyzes the definition of a bean and populates its autowire fields.
func analyzeDefinition(def *BeanDef) {
	for i := 0; i < def.OriginType.NumField(); i++ {
		field := def.OriginType.Field(i)
		if field.Anonymous {
			continue
		}

		valueTag, existValueTag := field.Tag.Lookup(injector.ValueTag)
		if existValueTag && valueTag != "" {
			// If the field has a value tag, it is treated as a config value injection point.
			// Config value injection over autowiring, so we skip autowire tag processing for this field.
			def.AutowireFields = append(def.AutowireFields, &AutowireField{
				Index:    i,
				Name:     field.Name,
				ValueTag: valueTag,
				Field:    field,
			})
			continue
		}

		// Process autowire tag for dependency injection.
		autowireTag, existAutowireTag := field.Tag.Lookup(injector.AutowireTag)
		if !existAutowireTag || autowireTag == "-" {
			continue
		}

		fieldKind := field.Type.Kind()
		isInterface := fieldKind == reflect.Interface

		if !isInterface && !(fieldKind == reflect.Pointer && field.Type.Elem().Kind() == reflect.Struct) {
			panic(fmt.Sprintf("field '%s' must be a pointer to a struct or an interface", field.Name))
		}

		def.AutowireFields = append(def.AutowireFields, &AutowireField{
			Index:       i,
			Name:        field.Name,
			IsInterface: isInterface,
			AutowireTag: autowireTag,
			Field:       field,
		})
	}
}

// createPrototypeBean creates a prototype bean instance
func createPrototypeBean(c *Container, def *BeanDef) any {
	beanValue := reflect.New(def.OriginType)

	err := doProcessFields(c, beanValue.Elem(), def.AutowireFields)
	if err != nil {
		panic(fmt.Sprintf("failed to initialize bean '%s': %s",
			def.OriginType.String(), err.Error()))
	}

	beanValueIf := beanValue.Interface()

	if postConstruct, ok := beanValueIf.(bean.PostConstruct); ok {
		postConstruct.BeanPostConstruct()
	}

	return beanValueIf
}

// doProcessFields processes the fields of a struct and injects dependencies based on autowire tags.
func doProcessFields(c *Container, structValue reflect.Value, autoFields []*AutowireField) error {
	for _, autoField := range autoFields {

		if autoField.ValueTag != "" {
			err := injector.WireConfigValue(structValue.Field(autoField.Index), autoField.Field.Type, autoField.ValueTag)
			if err != nil {
				return fmt.Errorf("wire config value for field '%s' failed: %s", autoField.Name, err.Error())
			}
			continue
		}

		var beanVal any
		var exist bool

		if autoField.IsInterface {
			prepareIFaceImplements(c, autoField.Field.Type)
		}

		if autoField.AutowireTag == "" {
			beanVal, exist = c.GetBeanByType(autoField.Field.Type)
		} else {
			beanVal, exist = c.GetBean(autoField.AutowireTag)
		}

		if !exist {
			return fmt.Errorf("no bean found for field '%s'", autoField.Name)
		}

		injector.WireBean(beanVal, structValue.Field(autoField.Index))
	}

	return nil
}

// initializeBeans initializes all beans in the container.
func initializeBeans(c *Container, ctrlType reflect.Type) {
	for beanName, beanDef := range c.beans {
		if beanDef.ready {
			continue
		}

		if postConstruct, ok := beanDef.Value.(bean.PostConstruct); ok {
			postConstruct.BeanPostConstruct()
		}

		if _, ok := beanDef.Value.(mvc.AbstractController); ok {
			c.typeMapping[ctrlType] = append(c.typeMapping[ctrlType], beanName)
		}

		beanDef.ready = true
	}
}

// prepareIFaceImplements prepares the interface implementations in the container.
func prepareIFaceImplements(c *Container, iface reflect.Type) {
	if c.LookupType(iface) {
		// If the interface type is already prepared, skip it.
		return
	}

	var implBeanNames []string

	for name, beanDef := range c.beans {
		if beanDef.Type.Implements(iface) {
			implBeanNames = append(implBeanNames, name)
		}
	}

	if len(implBeanNames) > 0 {
		c.typeMapping[iface] = implBeanNames
	}
}
