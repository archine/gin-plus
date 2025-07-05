package ioc

import (
	"errors"
	"fmt"
	"reflect"
	"sync"
	"sync/atomic"

	"github.com/archine/gin-plus/v4/util/strutil"
)

var (
	once            sync.Once
	refreshed       atomic.Bool
	beanType        = reflect.TypeOf((*Bean)(nil)).Elem()
	lazyBeanType    = reflect.TypeOf((*LazyBean)(nil)).Elem()
	ErrBeanNotFound = errors.New("bean not found")
)

var defaultContainer = &Container{
	beans:       make(map[string]any),
	definitions: make(map[string]*BeanDefinition),
	typeMapping: make(map[reflect.Type][]string),
}

// Container responsible for managing the lifecycle of all beans
type Container struct {
	mu          sync.RWMutex
	beans       map[string]any             // stores created bean instances
	definitions map[string]*BeanDefinition // stores bean definition information
	typeMapping map[reflect.Type][]string  // stores type to bean names mapping
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
		return nil, ErrBeanNotFound
	}

	if len(names) == 1 {
		c.mu.RLock()
		bean, exists := c.beans[names[0]]
		c.mu.RUnlock()

		if !exists {
			return nil, ErrBeanNotFound
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
		return nil, ErrBeanNotFound
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
		return nil, ErrBeanNotFound
	}
	return beans, nil
}

// registerBeanDefinitions handles bean registrations with validation and dependency analysis
func (c *Container) registerBeanDefinitions(structPtrTypes ...reflect.Type) error {
	for _, typ := range structPtrTypes {
		if typ == nil {
			continue
		}

		// validation
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
			if err := c.registerBeanDefinitions(newStructTypes...); err != nil {
				return err
			}
		}
	}
	return nil
}

// RegisterBean registers a bean instance to the IOC container.
// This method allows registration of pre-created objects that don't need to go through
// the automatic bean creation process.
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

// RegisterBeanDefinition registers bean definitions with the IOC container
func RegisterBeanDefinition(structPtrTypes ...reflect.Type) error {
	if len(structPtrTypes) == 0 {
		return nil
	}
	if refreshed.Load() {
		return errors.New("container already initialized, cannot register new bean definitions")
	}

	return defaultContainer.registerBeanDefinitions(structPtrTypes...)
}

// RegisterBean manually registers a bean instance to the IOC container.
// This method allows registration of pre-created objects that don't need to go through
// the automatic bean creation process.
//
// Parameters:
//   - name: bean name for registration. If empty, defaults to the struct name with first letter lowercase
//   - objPtr: pointer to the struct instance to register (must not be nil)
//   - implementedTypes: optional interface types that the object implements for type-based lookup
//
// Usage:
//
//	// Register with auto-generated name
//	ioc.RegisterBean("", &UserService{})
//
//	// Register with custom name and interface types
//	ioc.RegisterBean("userSvc", &UserService{}, reflectutil.InterfaceOf[UserRepository]())
//
// Notes:
//   - The object doesn't need to implement the Bean interface
//   - Only struct pointers are accepted
//   - If the object implements interfaces and you need type-based lookup, specify them in implementedTypes
//   - Registration will fail if a bean with the same name already exists
func RegisterBean(name string, objPtr any, implementedTypes ...reflect.Type) error {
	if objPtr == nil {
		return errors.New("objPtr cannot be nil")
	}

	return defaultContainer.RegisterBean(name, objPtr, implementedTypes...)
}

// GetBean retrieves a bean instance by its name.
// This method allows you to get a specific bean instance that has been registered in the IOC container
// Args:
//   - name: the name of the bean to retrieve
//
// Returns:
//   - the bean instance if found, or nil if not found
func GetBean(name string) (any, bool) {
	return defaultContainer.GetBean(name)
}

// GetBeanByType retrieves a bean instance by its type using generics.
//
// Args:
//   - T: the type of the bean to retrieve. This should be a struct or interface type, e.g. UserService
//
// Usage:
//
//	svc, err := ioc.GetBeanByType[UserService]()
//
// Returns:
//   - The bean instance as *T if found, or nil and an error if not found or type mismatch.
func GetBeanByType[T any]() (*T, error) {
	typ := reflect.TypeOf((*T)(nil)).Elem()

	bean, err := defaultContainer.GetBeanByType(typ)
	if err != nil {
		return nil, err
	}
	instance, ok := bean.(*T)
	if !ok {
		return nil, fmt.Errorf("bean is not of type %s", typ.String())
	}
	return instance, nil
}

// returns the default container instance.
//
// Note: this function is system-internal and should not be used directly in application code.
func getContainer() *Container {
	return defaultContainer
}
