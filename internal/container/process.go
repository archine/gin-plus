package container

import (
	"fmt"
	"strings"

	"reflect"

	"github.com/archine/gin-plus/v4/component/ioc/bean"
	"github.com/archine/gin-plus/v4/component/mvc"
	"github.com/archine/gin-plus/v4/internal/container/injector"
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

		// Parse optional flag: `autowire:"name,optional"` or `autowire:",optional"`
		beanName := autowireTag
		optional := false
		if before, after, ok := strings.Cut(autowireTag, ","); ok {
			beanName = strings.TrimSpace(before)
			option := strings.TrimSpace(after)
			if option != "" && option != "optional" {
				panic(fmt.Sprintf("field '%s' has unsupported autowire option '%s'; only 'optional' is supported", field.Name, option))
			}
			optional = option == "optional"
		}

		def.AutowireFields = append(def.AutowireFields, &AutowireField{
			Index:       i,
			Name:        field.Name,
			IsInterface: isInterface,
			AutowireTag: beanName,
			Optional:    optional,
			Field:       field,
		})
	}
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
			c.prepareIFaceImplements(autoField.Field.Type)
		}

		if autoField.AutowireTag == "" {
			beanVal, exist = c.GetBeanByType(autoField.Field.Type)
		} else {
			beanVal, exist = c.GetBean(autoField.AutowireTag)
		}

		if !exist {
			if autoField.Optional {
				continue
			}
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
func (c *Container) prepareIFaceImplements(iface reflect.Type) {
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
