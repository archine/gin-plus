package definition

import (
	"fmt"
	"github.com/archine/gin-plus/v4/util/strutil"
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
	Ready          bool
	Name           string             // name of the bean, defaults to struct name with first letter lowercase
	Prototype      bool               // whether this bean is a prototype (new instance each time) or singleton (shared instance)
	IsLaze         bool               // whether this bean is lazy-loaded (not instantiated until first requested)
	Bean           any                // actual bean instance, if available
	Type           reflect.Type       // type of the bean
	DependentBeans []string           // names of beans this bean depends on
	AutowireFields []*DependencyField // fields that require autowiring
}

// Cache holds the registered bean definitions
var Cache = make(map[string]*BeanDefinition)

// Process processes the given instance and registers its bean definition
func Process(instances ...any) error {
	for _, instance := range instances {
		if instance == nil {
			continue
		}

		typ := reflect.TypeOf(instances)

		if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Struct {
			return fmt.Errorf("type '%s' must be a pointer to a struct", typ.Name())
		}

		if instances

		if !typ.Implements(beanType) {
			return fmt.Errorf("type '%s' must implement Bean interface", typ.Name())
		}

		structType := typ.Elem()
		beanName := strutil.FirstToLower(structType.Name())

		// check if already registered
		if _, exists := c.definitions[beanName]; exists {
			return fmt.Errorf("bean '%s' already registered", beanName)
		}

		def := &BeanDefinition{
			Type: structType,
		}

		// analyze dependencies and register recursively
		newStructTypes, err := c.analyzeDependencies(def)
		if err != nil {
			return fmt.Errorf("failed to analyze dependencies for '%s': %w", beanName, err)
		}

		c.definitions[beanName] = def
		c.typeMapping[structType] = []string{beanName}

		// recursively register dependencies
		if len(newStructTypes) > 0 {
			if err := c.processStructType(newStructTypes); err != nil {
				return err
			}
		}
	}
	return nil
}
