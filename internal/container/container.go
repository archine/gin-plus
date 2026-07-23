package container

import (
	"fmt"
	"reflect"
	"sync"

	"github.com/archine/gin-plus/v4/component/config"
	"github.com/archine/gin-plus/v4/component/mvc"
)

// AutowireField represents a field metadata for dependency injection.
type AutowireField struct {
	Index       int                 // Field index in the struct
	IsInterface bool                // Whether the field is an interface
	Name        string              // Field name
	ValueTag    string              // Value from `value` tag (config)
	AutowireTag string              // Value from `autowire` tag (bean name)
	Optional    bool                // If true, no error is thrown if bean is missing
	Field       reflect.StructField // Original reflect field
}

// BeanDef represents the definition and lifecycle state of a bean.
type BeanDef struct {
	ready          bool                       // Initialization status
	Condition      func(config.Provider) bool // Optional refresh-time registration guard
	Value          any                        // Actual instance (must be a pointer)
	Type           reflect.Type               // Full type information (e.g., *UserService)
	OriginType     reflect.Type               // The underlying struct type (e.g., UserService)
	AutowireFields []*AutowireField           // Fields identified for injection
}

// Container manages the registration, dependency injection, and lifecycle of beans.
type Container struct {
	mu   sync.RWMutex
	once sync.Once

	// beans stores definitions indexed by unique bean names.
	beans map[string]*BeanDef

	// typeMapping stores aliases for types (structs and interfaces) to bean names.
	typeMapping map[reflect.Type][]string
}

// NewContainer initializes a new IOC container.
func NewContainer() *Container {
	return &Container{
		beans:       make(map[string]*BeanDef),
		typeMapping: make(map[reflect.Type][]string),
	}
}

// LookupType checks if a specific type (or its pointer element) is registered.
func (c *Container) LookupType(typ reflect.Type) bool {
	if typ == nil {
		return false
	}
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	c.mu.RLock()
	defer c.mu.RUnlock()
	_, exists := c.typeMapping[typ]
	return exists
}

// GetBean retrieves a bean instance by its registered name.
func (c *Container) GetBean(name string) (any, bool) {
	if name == "" {
		return nil, false
	}

	c.mu.RLock()
	defer c.mu.RUnlock()

	def, exists := c.beans[name]
	if !exists {
		return nil, false
	}

	return def.Value, true
}

// GetBeanByType retrieves a single bean instance by its type.
// If multiple candidates exist for the same type, it will panic to avoid ambiguity.
func (c *Container) GetBeanByType(typ reflect.Type) (any, bool) {
	if typ == nil {
		return nil, false
	}
	// Normalize pointers for struct lookup
	searchTyp := typ
	if searchTyp.Kind() == reflect.Pointer {
		searchTyp = searchTyp.Elem()
	} else if searchTyp.Kind() == reflect.Interface {
		// For interfaces, we want to search by the interface type itself
		c.prepareIFaceImplements(searchTyp)
	}

	c.mu.RLock()
	names, exists := c.typeMapping[searchTyp]
	c.mu.RUnlock()

	if !exists || len(names) == 0 {
		return nil, false
	}

	if len(names) > 1 {
		panic(fmt.Sprintf("[IOC] ambiguous dependency: multiple beans found for type %v: %v", typ, names))
	}

	return c.GetBean(names[0])
}

// GetAllBeansByType retrieves all bean instances that match the specified type or interface.
func (c *Container) GetAllBeansByType(typ reflect.Type) ([]any, bool) {
	if typ == nil {
		return nil, false
	}
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	} else if typ.Kind() == reflect.Interface {
		// For interfaces, we want to search by the interface type itself
		c.prepareIFaceImplements(typ)
	}

	c.mu.RLock()
	names, exists := c.typeMapping[typ]
	c.mu.RUnlock()

	if !exists || len(names) == 0 {
		return nil, false
	}

	beans := make([]any, 0, len(names))
	for _, name := range names {
		if val, ok := c.GetBean(name); ok {
			beans = append(beans, val)
		}
	}

	return beans, len(beans) > 0
}

// RegisterBean registers a pre-instantiated object as a singleton bean.
// It allows mapping the instance to specific interface types for dependency injection.
// If no explicit name is provided, the container uses package.structName as the default.
func (c *Container) RegisterBean(name string, instance any, itypes ...reflect.Type) {
	if instance == nil {
		panic("[IOC] registration failed: instance cannot be empty")
	}

	beanTyp := reflect.TypeOf(instance)
	if beanTyp.Kind() != reflect.Pointer || beanTyp.Elem().Kind() != reflect.Struct {
		panic(fmt.Sprintf("[IOC] registration failed: expected struct pointer, got %T", instance))
	}

	structType := beanTyp.Elem()
	if name == "" {
		name = structType.String()
	}

	c.mu.Lock()
	defer c.mu.Unlock()

	if old, exists := c.beans[name]; exists {
		panic(fmt.Sprintf("[IOC] bean name conflict: '%s' is already registered by %v", name, old.Type))
	}

	// Register as a ready singleton
	c.beans[name] = &BeanDef{
		ready:      true,
		Value:      instance,
		Type:       beanTyp,
		OriginType: structType,
	}

	// Map to its own struct type
	c.typeMapping[structType] = append(c.typeMapping[structType], name)

	// Map to provided interfaces with strict implementation check
	for _, itype := range itypes {
		if !beanTyp.Implements(itype) {
			continue
		}
		c.typeMapping[itype] = append(c.typeMapping[itype], name)
	}

	// Automatic MVC controller detection
	if _, ok := instance.(mvc.AbstractController); ok {
		ctrlType := reflect.TypeFor[mvc.AbstractController]()
		c.typeMapping[ctrlType] = append(c.typeMapping[ctrlType], name)
	}
}

// Refresh builds the effective bean graph and triggers dependency injection.
// It should be invoked after configuration is initialized and all bean definitions are registered.
func (c *Container) Refresh(cp config.Provider, registry *BeanDefinitionRegistry) {
	if cp == nil {
		panic("[IOC] config provider is not initialized before container refresh")
	}
	if registry == nil {
		panic("[IOC] bean definition registry is not initialized")
	}

	c.once.Do(func() {
		defs := registry.DrainDefs()

		for beanName, beanDef := range defs {
			if beanDef.Condition != nil && !beanDef.Condition(cp) {
				delete(defs, beanName)
				continue
			}

			c.mu.Lock()
			c.registerBeanDefLocked(beanName, beanDef)
			c.mu.Unlock()
		}

		ctrlType := reflect.TypeFor[mvc.AbstractController]()
		for beanName, beanDef := range defs {
			// Parse tags and prepare AutowireFields
			analyzeDefinition(beanDef)

			// Process fields and satisfy dependencies
			targetVal := reflect.ValueOf(beanDef.Value).Elem()
			err := doProcessFields(c, targetVal, beanDef.AutowireFields)
			if err != nil {
				panic(fmt.Sprintf("[IOC] failed to initialize bean '%s' (%v): %v", beanName, beanDef.Type, err))
			}
		}

		// Perform additional initialization (e.g., MVC routing)
		initializeBeans(c, ctrlType)
	})
}

func (c *Container) registerBeanDefLocked(name string, def *BeanDef) {
	if old, exists := c.beans[name]; exists {
		panic(fmt.Sprintf("[IOC] bean name conflict: '%s' is already registered as %v, cannot register %v",
			name, old.Type, def.Type))
	}
	c.beans[name] = def
	c.typeMapping[def.OriginType] = append(c.typeMapping[def.OriginType], name)
}
