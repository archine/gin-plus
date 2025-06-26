package container

import (
	"fmt"
	"reflect"

	"github.com/archine/gin-plus/v4/component/bean"
	"github.com/archine/gin-plus/v4/internal/container/topo"
)

// getStructDependsBean analyzes struct fields to extract dependency information
func getStructDependsBean(def *beanDefinition) ([]reflect.Type, error) {
	var notRegisterStruct []reflect.Type

	var fieldKind reflect.Kind
	var fieldIsInterface bool
	var fieldOriginType reflect.Type

	for i := 0; i < def.typ.NumField(); i++ {
		field := def.typ.Field(i)
		if field.Anonymous {
			continue
		}

		autowire := field.Tag.Get("autowire")
		if autowire == "" {
			continue
		}

		fieldKind = field.Type.Kind()
		fieldIsInterface = fieldKind == reflect.Interface
		fieldOriginType = field.Type

		// validate field type: must be interface or pointer to struct
		if !fieldIsInterface && !(fieldKind == reflect.Ptr && fieldOriginType.Elem().Kind() == reflect.Struct) {
			return nil, fmt.Errorf("field '%s' in '%s' must be interface or struct pointer", field.Name, def.typ.Name())
		}

		if fieldKind == reflect.Ptr {
			fieldOriginType = field.Type.Elem()
		}

		def.dependFields = append(def.dependFields, &structField{
			idx:         i,
			name:        fieldOriginType.Name(),
			isInterface: fieldIsInterface,
			autowire:    autowire,
			field:       field,
		})

		// resolve bean definition for the field type
		if names, exists := container.typeForNames[fieldOriginType]; exists {
			def.dependBeanNames = append(def.dependBeanNames, names...)
		} else {
			// handle interface types by finding implementing beans
			if fieldIsInterface {
				implementingBeans := findBeansImplementingInterface(field.Type)
				if len(implementingBeans) == 0 {
					return nil, fmt.Errorf("no implementation found for interface '%s'", fieldOriginType.String())
				}
				def.dependBeanNames = append(def.dependBeanNames, implementingBeans...)
				// cache the result to avoid repeated lookups
				container.typeForNames[fieldOriginType] = implementingBeans
			} else {
				if field.Type.Implements(componentType) {
					notRegisterStruct = append(notRegisterStruct, field.Type)
				} else {
					return nil, fmt.Errorf("field '%s' in '%s' must implement bean.Marker", field.Name, def.typ.Name())
				}
			}
		}
	}

	return notRegisterStruct, nil
}

// buildBeanCreatOrder builds bean creation order using topological sorting
// It returns a sorted list of bean names that should be created in order
func buildBeanCreatOrder() ([]string, error) {
	var edges []*topo.DependencyEdge

	for beanName, def := range container.befCache {
		// create dependency edges: beanName depends on dependency
		for _, dependency := range def.dependBeanNames {
			edge := &topo.DependencyEdge{
				From: beanName,   // dependent bean
				To:   dependency, // dependency bean (must be created first)
			}
			edges = append(edges, edge)
		}
	}

	if len(edges) == 0 {
		return nil, nil
	}

	return topo.Sort(edges)
}

// findBeansImplementingInterface finds all beans that implement the specified interface
func findBeansImplementingInterface(interfaceType reflect.Type) []string {
	implementingBeans := make([]string, 0, 2)

	for beanName, def := range container.befCache {
		if def.typ.Implements(interfaceType) {
			implementingBeans = append(implementingBeans, beanName)
		}
	}

	return implementingBeans
}

// createBean creates a bean instance and performs dependency injection
func createBean(structName string) error {
	if _, exists := container.beans[structName]; exists {
		return nil // already created
	}

	def, exists := container.befCache[structName]
	if !exists {
		return fmt.Errorf("bean definition not found for '%s'", structName)
	}

	beanValue := reflect.New(def.typ).Elem()

	for _, sf := range def.dependFields {
		if sf.autowire != "-" {
			// look up the dependency bean by its name
			dependencyBean, exist := container.beans[sf.autowire]
			if !exist {
				return fmt.Errorf("dependency bean '%s' not found", sf.autowire)
			}

			dependencyType := reflect.TypeOf(dependencyBean)

			if sf.isInterface {
				if dependencyType.Implements(sf.field.Type) {
					beanValue.Field(sf.idx).Set(reflect.ValueOf(dependencyBean))
					continue
				}
				return fmt.Errorf("bean '%s' does not implement interface '%s'", sf.autowire, sf.field.Type.String())
			}

			if dependencyType.AssignableTo(sf.field.Type) {
				beanValue.Field(sf.idx).Set(reflect.ValueOf(dependencyBean))
				continue
			}

			return fmt.Errorf("type mismatch: expected '%s', got '%s'", sf.field.Type.String(), dependencyType.String())
		}
	}

	beanInstance := beanValue.Interface()

	// call post-construct method if implemented
	if pc, ok := beanInstance.(bean.PostConstruct); ok {
		pc.BeanPostConstruct()
	}

	container.beans[structName] = beanInstance
	return nil
}
