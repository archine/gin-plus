package container

import (
	"fmt"
	"reflect"

	"github.com/archine/gin-plus/v4/component/ioc"
	"github.com/archine/gin-plus/v4/internal/container/injection"
	"github.com/archine/gin-plus/v4/internal/container/registry"
	"github.com/archine/gin-plus/v4/util/strutil"
)

var (
	beanType = reflect.TypeOf((*ioc.AbstractBean)(nil)).Elem()
)

// processDefinitions processes all bean definition in the container.
// It analyzes each definition to find dependencies and autowire fields.
func processDefinitions(c *Container) error {
	for registry.HasNext() {
		def := registry.Pop()
		if def == nil {
			continue
		}

		if bean, exist := c.beans[def.Name]; exist {
			// duplicate bean definition, skip it
			return fmt.Errorf("duplicate bean name '%s' detected: [%s, %s]. please ensure each bean has a unique name",
				def.Name, bean.originTyp.String(), def.OriginType.String())
		}

		err := analyzeStruct(c, def)
		if err != nil {
			return err
		}

		c.typeMapping[def.OriginType] = []string{def.Name}
		c.beans[def.Name] = &BeanDef{
			originTyp:      def.OriginType,
			isPrototype:    def.IsPrototype,
			value:          def.Bean,
			autowireFields: def.AutowireFields,
		}
	}

	beanType = nil
	registry.Clean()

	return nil
}

// analyzeStruct analyzes a bean definition to find its dependencies and autowire fields.
func analyzeStruct(c *Container, def *registry.BeanDefinition) error {
	for i := 0; i < def.OriginType.NumField(); i++ {
		field := def.OriginType.Field(i)
		if field.Anonymous {
			continue
		}

		valueTag, ok := field.Tag.Lookup(injection.ValueTag)
		if ok && valueTag != "" {
			// If the field has a configuration tag, it is not an autowire field.
			// We can skip it for autowiring.
			def.AutowireFields = append(def.AutowireFields, &registry.AutowireField{
				Index:    i,
				Name:     field.Name,
				ValueTag: valueTag,
				Field:    field,
			})
			continue
		}

		autowireTag := field.Tag.Get(injection.AutowireTag)
		if autowireTag == "" {
			continue
		}

		fieldType := field.Type

		if registry.LookupType(fieldType) {
			// If the field type is already registered, skip it.
			continue
		}

		autoField := registry.AutowireField{
			Index:       i,
			Name:        field.Name,
			AutowireTag: autowireTag,
			Field:       field,
		}

		if fieldType.Kind() == reflect.Interface {
			autoField.IsInterface = true
			implBeanNames, exist := c.typeMapping[fieldType]

			if !exist {
				implBeanNames = findImplBeanNames(fieldType)
				if len(implBeanNames) == 0 {
					continue
				}
			}
			c.typeMapping[fieldType] = implBeanNames

		} else {
			if fieldType.Kind() != reflect.Ptr {
				return fmt.Errorf("field '%s' in '%s' must be a pointer type", field.Name, def.OriginType.String())
			}

			fieldOriginType := fieldType.Elem()

			if fieldOriginType == def.OriginType {
				return fmt.Errorf("field '%s' (type: %s) in '%s' cannot depend on itself (circular dependency)",
					field.Name, fieldType.String(), def.OriginType.String())
			}

			if fieldType.Implements(beanType) {
				fieldInstance := reflect.New(fieldOriginType).Interface()
				ib, _ := fieldInstance.(ioc.AbstractBean)

				newDef := &registry.BeanDefinition{
					Bean:        fieldInstance,
					Name:        ib.BeanName(),
					PtrType:     fieldType,
					OriginType:  fieldOriginType,
					IsPrototype: ib.IsPrototype(),
				}
				if newDef.Name == "" {
					newDef.Name = strutil.FirstToLower(fieldOriginType.Name())
				}

				registry.RegisterBeanDefinition(newDef)

			} else {
				return fmt.Errorf("field '%s' in struct '%s' is not declared as a Bean and cannot be automatically registered as a BeanDefinition",
					field.Name, def.OriginType.String())
			}
		}

		def.AutowireFields = append(def.AutowireFields, &autoField)
	}

	return nil
}

// findImplBeanNames finds all bean names that implement the given interface type.
func findImplBeanNames(interfaceType reflect.Type) []string {
	var implBeanNames []string

	allDefs := registry.GetAllDefinitions()
	for _, def := range allDefs {
		if def.PtrType.Implements(interfaceType) {
			implBeanNames = append(implBeanNames, def.Name)
		}
	}

	return implBeanNames
}
