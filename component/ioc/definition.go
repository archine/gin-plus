package ioc

import (
	"fmt"
	"github.com/archine/gin-plus/v4/internal/container/definition"
	"github.com/archine/gin-plus/v4/util/strutil"
	"reflect"
)

// RegisterBeanDefinition registers a bean definition in the IoC container.
// This function is used to register bean definitions that will be managed by the IoC container.
// It accepts one or more pointers to structs that implement the Bean interface.
//
// Args:
//   - instances: one or more struct pointers that implement the Bean interface.
func RegisterBeanDefinition(instances ...any) error {
	if len(instances) == 0 {
		return nil
	}

	for _, instance := range instances {
		if instance == nil {
			continue
		}

		ib, ok := instance.(Bean)
		if !ok {
			continue
		}

		typ := reflect.TypeOf(instance)
		if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Struct {
			return fmt.Errorf("type '%s' must be a pointer to a struct", typ.Name())
		}

		typ = typ.Elem()

		beanName := ib.BeanName()
		if beanName == "" {
			beanName = strutil.FirstToLower(typ.Name())
		}

		if _, found := definition.Cache[beanName]; found {
			return fmt.Errorf("bean '%s' already registered", beanName)
		}

		def := &definition.BeanDefinition{
			Name:           beanName,
			Type:           typ,
			Prototype:      ib.IsPrototype(),
			IsLaze:         ib.IsLazy(),
			DependentBeans: make([]string, 0, typ.NumField()),
			AutowireFields: make([]*definition.DependencyField, 0, typ.NumField()),
			Bean:           instance,
		}

		definition.Cache[beanName] = def
	}

	return nil
}

// analyzeDependencies analyzes struct fields to extract dependency information
//func analyzeDependencies(def *definition.BeanDefinition) ([]reflect.Type, error) {
//	var unregisteredTypes []reflect.Type
//
//	for i := 0; i < def.Type.NumField(); i++ {
//		field := def.Type.Field(i)
//		if field.Anonymous {
//			continue
//		}
//
//		autowireTag := field.Tag.Get("autowire")
//		if autowireTag == "" {
//			continue
//		}
//
//		fieldType := field.Type
//		isInterface := fieldType.Kind() == reflect.Interface
//		isPointer := fieldType.Kind() == reflect.Ptr
//
//		if !isInterface && !(isPointer && fieldType.Elem().Kind() == reflect.Struct) {
//			return nil, fmt.Errorf("field '%s' in '%s' must be interface or struct pointer",
//				field.Name, def.Type.Name())
//		}
//
//		targetType := fieldType
//		if isPointer {
//			targetType = fieldType.Elem()
//		}
//
//		if targetType == def.Type {
//			return nil, fmt.Errorf("field '%s' in '%s' cannot depend on itself",
//				field.Name, def.Type.Name())
//		}
//
//		if names, exists := c.typeMapping[targetType]; exists {
//			// if we already have registered beans for this type, use them
//			if len(names) == 0 {
//				return nil, fmt.Errorf("no beans registered for type '%s'", targetType.String())
//			}
//			// append existing bean names to dependent beans
//			def.DependentBeans = append(def.DependentBeans, names...)
//		} else {
//			if isInterface {
//				implementingBeans := c.findImplementingBeans(fieldType)
//				if len(implementingBeans) == 0 {
//					return nil, fmt.Errorf("no implementation found for interface '%s'",
//						targetType.String())
//				}
//				def.DependentBeans = append(def.DependentBeans, implementingBeans...)
//				c.typeMapping[targetType] = implementingBeans
//			} else {
//				// for struct types, we need to register them if not already done
//				if fieldType.Implements(beanType) {
//					unregisteredTypes = append(unregisteredTypes, fieldType)
//				} else {
//					return nil, fmt.Errorf("field '%s' in '%s' must implement Component interface",
//						field.Name, def.Type.Name())
//				}
//			}
//		}
//
//		def.DependentFields = append(def.DependentFields, &DependencyField{
//			Index:       i,
//			Name:        targetType.Name(),
//			IsInterface: isInterface,
//			AutowireTag: autowireTag,
//			Field:       field,
//		})
//	}
//
//	return unregisteredTypes, nil
//}
//
//// findImplementingBeans finds all beans that implement the specified interface
//func findImplementingBeans(interfaceType reflect.Type) []string {
//	var implementingBeans []string
//
//	for beanName, def := range c.definitions {
//		if def.Type.Implements(interfaceType) {
//			implementingBeans = append(implementingBeans, beanName)
//		}
//	}
//
//	return implementingBeans
//}
