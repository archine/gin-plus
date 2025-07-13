package container

import (
	"fmt"
	"github.com/archine/gin-plus/v4/component/ioc"
	"github.com/archine/gin-plus/v4/internal/container/definition"
	"github.com/archine/gin-plus/v4/util/strutil"
	"reflect"
)

var (
	beanType = reflect.TypeOf((*ioc.AbstractBean)(nil)).Elem()
)

// processDefinitions processes all bean definition in the container.
// It analyzes each definition to find dependencies and autowire fields.
func processDefinitions(c *Container) error {
	for definition.HasNext() {
		def := definition.Pop()
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
	definition.Clean()

	return nil
}

// analyzeStruct analyzes a bean definition to find its dependencies and autowire fields.
func analyzeStruct(c *Container, def *definition.BeanDefinition) error {
	for i := 0; i < def.OriginType.NumField(); i++ {
		field := def.OriginType.Field(i)
		if field.Anonymous {
			continue
		}

		autowireTag := field.Tag.Get(definition.AutowireTag)
		if autowireTag == "" {
			continue
		}

		fieldType := field.Type

		autoField := definition.AutowireField{
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
					return fmt.Errorf("no definitions found for field '%s' in '%s'", field.Name, def.OriginType.String())
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

			if !definition.LookupType(fieldOriginType) {
				if fieldType.Implements(beanType) {
					fieldInstance := reflect.New(fieldOriginType).Interface()
					ib, _ := fieldInstance.(ioc.AbstractBean)

					newDef := &definition.BeanDefinition{
						Name:        ib.BeanName(),
						IsPrototype: ib.IsPrototype(),
						PtrType:     fieldType,
						OriginType:  fieldOriginType,
						Bean:        fieldInstance,
					}
					if newDef.Name == "" {
						newDef.Name = strutil.FirstToLower(fieldOriginType.Name())
					}

					definition.RegisterBeanDefinition(newDef)
				} else {
					return fmt.Errorf("field '%s' in '%s' is not a bean and cannot be auto-registered",
						field.Name, def.OriginType.String())
				}
			}
		}

		def.AutowireFields = append(def.AutowireFields, &autoField)
	}

	return nil
}

// findImplBeanNames finds all bean names that implement the given interface type.
func findImplBeanNames(interfaceType reflect.Type) []string {
	var implBeanNames []string

	allDefs := definition.GetAllDefinitions()
	for _, def := range allDefs {
		if def.PtrType.Implements(interfaceType) {
			implBeanNames = append(implBeanNames, def.Name)
		}
	}

	return implBeanNames
}
