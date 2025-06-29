package ioc

import (
	"fmt"
	"reflect"
)

// DependencyField represents dependency injection information for struct fields
type DependencyField struct {
	Index       int                 // field index in struct
	Name        string              // field name
	IsInterface bool                // whether field is interface type
	AutowireTag string              // autowire tag value
	Field       reflect.StructField // field reflection info
}

// BeanDefinition represents bean definition information
type BeanDefinition struct {
	Type           reflect.Type       // bean type
	Dependencies   []*DependencyField // dependency fields
	DependentBeans []string           // names of dependent beans
}

// analyzeDependencies analyzes struct fields to extract dependency information
func (c *Container) analyzeDependencies(def *BeanDefinition) ([]reflect.Type, error) {
	var unregisteredTypes []reflect.Type

	for i := 0; i < def.Type.NumField(); i++ {
		field := def.Type.Field(i)
		if field.Anonymous {
			continue
		}

		autowireTag := field.Tag.Get("autowire")
		if autowireTag == "" {
			continue
		}

		fieldType := field.Type
		isInterface := fieldType.Kind() == reflect.Interface
		isPointer := fieldType.Kind() == reflect.Ptr

		// validate field type
		if !isInterface && !(isPointer && fieldType.Elem().Kind() == reflect.Struct) {
			return nil, fmt.Errorf("field '%s' in '%s' must be interface or struct pointer",
				field.Name, def.Type.Name())
		}

		targetType := fieldType
		if isPointer {
			targetType = fieldType.Elem()
		}

		depField := &DependencyField{
			Index:       i,
			Name:        targetType.Name(),
			IsInterface: isInterface,
			AutowireTag: autowireTag,
			Field:       field,
		}

		def.Dependencies = append(def.Dependencies, depField)

		if names, exists := c.typeMapping[targetType]; exists {
			// if we already have registered beans for this type, use them
			if len(names) == 0 {
				return nil, fmt.Errorf("no beans registered for type '%s'", targetType.String())
			}
			// append existing bean names to dependent beans
			def.DependentBeans = append(def.DependentBeans, names...)
		} else {
			if isInterface {
				implementingBeans := c.findImplementingBeans(fieldType)
				if len(implementingBeans) == 0 {
					return nil, fmt.Errorf("no implementation found for interface '%s'",
						targetType.String())
				}
				def.DependentBeans = append(def.DependentBeans, implementingBeans...)
				c.typeMapping[targetType] = implementingBeans
			} else {
				// for struct types, we need to register them if not already done
				if fieldType.Implements(beanType) {
					unregisteredTypes = append(unregisteredTypes, fieldType)
				} else {
					return nil, fmt.Errorf("field '%s' in '%s' must implement Component interface",
						field.Name, def.Type.Name())
				}
			}
		}
	}

	return unregisteredTypes, nil
}

// findImplementingBeans finds all beans that implement the specified interface
func (c *Container) findImplementingBeans(interfaceType reflect.Type) []string {
	var implementingBeans []string

	for beanName, def := range c.definitions {
		if def.Type.Implements(interfaceType) {
			implementingBeans = append(implementingBeans, beanName)
		}
	}

	return implementingBeans
}
