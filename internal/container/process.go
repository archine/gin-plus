package container

import (
	"fmt"
	"github.com/archine/gin-plus/v4/component/ioc"
	"github.com/archine/gin-plus/v4/internal/container/definition"
	"github.com/archine/gin-plus/v4/internal/container/topo"
	"github.com/archine/gin-plus/v4/util/strutil"
	"reflect"
)

var (
	// beanType is the type of the Bean interface
	beanType = reflect.TypeOf((*ioc.Bean)(nil)).Elem()
)

// processDefinition processes all bean definitions in the container.
// It analyzes each definition to find dependencies and autowire fields.
// It returns a sorted list of bean names in the order they should be created.
func (c *Container) processDefinition() ([]string, error) {
	var allEdges []*topo.Edge

	for _, def := range definition.Cache {
		edges, err := c.analyze(def)
		if err != nil {
			return nil, err
		}
		if len(edges) > 0 {
			allEdges = append(allEdges, edges...)
		}
	}

	if len(allEdges) == 0 {
		return topo.Sort(allEdges)
	}

	return nil, nil
}

// analyze analyzes a bean definition to find its dependencies and autowire fields.
// It returns a slice of edges representing the dependencies of the bean.
func (c *Container) analyze(def *definition.BeanDefinition) ([]*topo.Edge, error) {
	var edges []*topo.Edge

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
				field.Name, def.Name)
		}

		targetType := fieldType
		if isPointer {
			targetType = fieldType.Elem()
		}

		if targetType == def.Type {
			return nil, fmt.Errorf("field '%s' in '%s' cannot depend on itself", field.Name, def.Type.Name())
		}

		if names, exists := c.typeMapping[targetType]; exists {
			// if we already have registered beans for this type, use them
			if len(names) == 0 {
				return nil, fmt.Errorf("no beans registered for type '%s'", targetType.String())
			}
			edges = append(edges, topo.BuildEdge(def.Name, names)...)
		} else {
			if isInterface {
				beanNames := c.findImplementingBeanNames(fieldType)
				if len(beanNames) == 0 {
					return nil, fmt.Errorf("no implementation found for interface '%s'",
						targetType.String())
				}
				edges = append(edges, topo.BuildEdge(def.Name, beanNames)...)
				c.typeMapping[targetType] = beanNames
			} else {
				// for struct types, we need to register them if not already done
				if fieldType.Implements(beanType) {
					fieldInstance := reflect.New(targetType).Interface()
					ib, _ := fieldInstance.(ioc.Bean)

					filedDef := &definition.BeanDefinition{
						Name:        ib.BeanName(),
						IsPrototype: ib.IsPrototype(),
						Type:        targetType,
						Bean:        fieldInstance,
					}
					if filedDef.Name == "" {
						filedDef.Name = strutil.FirstToLower(targetType.Name())
					}

					definition.Cache[filedDef.Name] = filedDef

					c.typeMapping[targetType] = []string{filedDef.Name}
					edges = append(edges, topo.BuildEdge(def.Name, []string{filedDef.Name})...)
				} else {
					return nil, fmt.Errorf("field '%s' in '%s' must implement Bean interface",
						field.Name, def.Type.Name())
				}
			}
		}

		def.AutowireFields = append(def.AutowireFields, &definition.AutowireField{
			Index:       i,
			Name:        targetType.Name(),
			IsInterface: isInterface,
			AutowireTag: autowireTag,
			Field:       field,
		})
	}

	return edges, nil
}

// findImplementingBeanNames finds all bean names that implement the given interface type.
func (c *Container) findImplementingBeanNames(interfaceType reflect.Type) []string {
	var implementingBeans []string

	for beanName, def := range definition.Cache {
		if def.Type.Implements(interfaceType) {
			implementingBeans = append(implementingBeans, beanName)
		}
	}

	return implementingBeans
}
