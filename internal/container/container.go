package container

import (
	"errors"
	"fmt"
	"github.com/archine/gin-plus/v4/component/ioc"
	"github.com/archine/gin-plus/v4/util/strutil"
	"reflect"
	"sync"
)

var (
	once         sync.Once
	beanType     = reflect.TypeOf((*ioc.Bean)(nil)).Elem()
	lazyBeanType = reflect.TypeOf((*ioc.LazyBean)(nil)).Elem()
)

// Container responsible for managing the lifecycle of all beans
type Container struct {
	mu          sync.RWMutex
	beans       map[string]any             // stores created bean instances
	definitions map[string]*BeanDefinition // stores bean definition information
	typeMapping map[reflect.Type][]string  // stores type to bean names mapping
}

func NewContainer() *Container {
	return &Container{
		beans:       make(map[string]any),
		definitions: make(map[string]*BeanDefinition),
		typeMapping: make(map[reflect.Type][]string),
	}
}

// GetBean gets bean instance by bean name
func (c *Container) GetBean(name string) (any, bool) {
	if name == "" {
		return nil, false
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	bean, exists := c.beans[name]
	return bean, exists
}

// GetBeanByType gets bean instance by type
func (c *Container) GetBeanByType(typ reflect.Type) (any, error) {
	if typ == nil {
		return nil, fmt.Errorf("type cannot be nil")
	}
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	c.mu.RLock()
	names, exists := c.typeMapping[typ]
	c.mu.RUnlock()

	if !exists || len(names) == 0 {
		return nil, fmt.Errorf("no beans found for type '%s'", typ.Name())
	}

	if len(names) == 1 {
		c.mu.RLock()
		bean, exists := c.beans[names[0]]
		c.mu.RUnlock()

		if !exists {
			return nil, fmt.Errorf("bean '%s' not found in container", names[0])
		}
		return bean, nil
	}

	return nil, fmt.Errorf("multiple beans found for type '%s': %v. "+
		"Consider using GetBean() with specific bean name instead",
		typ.Name(), names)
}

// GetAllBeansByType retrieves all beans that implement the specified type
func (c *Container) GetAllBeansByType(typ reflect.Type) ([]any, error) {
	if typ == nil {
		return nil, fmt.Errorf("type cannot be nil")
	}
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	c.mu.RLock()
	names, exists := c.typeMapping[typ]
	c.mu.RUnlock()
	if !exists || len(names) == 0 {
		return nil, fmt.Errorf("no beans found for type '%s'", typ.Name())
	}

	beans := make([]any, 0, len(names))
	c.mu.RLock()
	for _, name := range names {
		if bean, exists := c.beans[name]; exists {
			beans = append(beans, bean)
		}
	}
	c.mu.RUnlock()

	if len(beans) == 0 {
		return nil, fmt.Errorf("no beans found for type '%s'", typ.Name())
	}
	return beans, nil
}

// RegisterBean manually registers a bean instance to the IOC container.
// This method allows registration of pre-created objects that don't need to go through
// the automatic bean creation process.
//
// Args:
//   - name: bean name for registration. If empty, defaults to the struct name with first letter lowercase
//   - objPtr: pointer to the struct instance to register (must not be nil)
//   - implementedTypes: optional interface types that the object implements for type-based lookup
func (c *Container) RegisterBean(name string, objPtr any, implementedTypes ...reflect.Type) error {
	objType := reflect.TypeOf(objPtr)
	if objType.Kind() != reflect.Ptr || objType.Elem().Kind() != reflect.Struct {
		return errors.New("objPtr must be a pointer to a struct")
	}
	structType := objType.Elem()
	if name == "" {
		name = strutil.FirstToLower(structType.Name())
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.beans[name]; exists {
		return fmt.Errorf("bean '%s' already exists", name)
	}

	// register implemented interfaces
	for _, iType := range implementedTypes {
		if !objType.Implements(iType) {
			return fmt.Errorf("object type '%s' does not implement interface '%s'",
				objType.Name(), iType.Name())
		}
		c.typeMapping[iType] = append(c.typeMapping[iType], name)
	}

	c.beans[name] = objPtr
	if _, exists := c.typeMapping[structType]; !exists {
		c.typeMapping[structType] = []string{name}
	}

	return nil
}

// processStructType processes a slice of struct pointer types to register them as definitions
func (c *Container) processStructType(structPtrTypes []reflect.Type) error {
	for _, typ := range structPtrTypes {
		if typ == nil {
			continue
		}

		if typ.Kind() != reflect.Ptr || typ.Elem().Kind() != reflect.Struct {
			return fmt.Errorf("type '%s' must be a pointer to a struct", typ.Name())
		}

		if typ.Implements(lazyBeanType) {
			continue // skip lazy init types
		}

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
