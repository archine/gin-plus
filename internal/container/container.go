package container

import (
	"fmt"
	"github.com/archine/gin-plus/v4/internal/container/injector"
	"github.com/archine/gin-plus/v4/util/strutil"
	"reflect"
	"sync"

	"github.com/archine/gin-plus/v4/component/mvc"
)

// AutowireField represents a field that requires autowiring
type AutowireField struct {
	// Index is the index of the field in the struct.
	Index int

	// IsInterface indicates whether the field is an interface type.
	IsInterface bool

	// Name is the name of the field.
	Name string

	// ValueTag value tag value, used for config values
	ValueTag string

	// AutowireTag is the tag used to autowire the field.
	AutowireTag string

	// Field is the reflect.StructField representing the field.
	Field reflect.StructField
}

// BeanDef represents a bean definition in the IOC container.
type BeanDef struct {
	// ready indicates whether the bean is ready for use.
	ready bool

	// Value is the actual bean instance.
	Value any

	// type is the type of the bean instance.
	Type reflect.Type

	// originType is the original type of the bean instance.
	// It is used to create new instances for prototype beans.
	OriginType reflect.Type

	// IsPrototype indicates whether the bean is a prototype (new instance for each request)
	IsPrototype bool

	// AutowireFields contains fields that need to be autowired.
	AutowireFields []*AutowireField
}

// Container responsible for managing the lifecycle of all beans
type Container struct {
	mu   sync.RWMutex
	once sync.Once

	// beans stores all registered beans by their names.
	beans map[string]*BeanDef

	// typeMapping maps types to their corresponding bean names for type-based retrieval.
	typeMapping map[reflect.Type][]string
}

func NewContainer() *Container {
	return &Container{
		beans:       make(map[string]*BeanDef),
		typeMapping: make(map[reflect.Type][]string),
	}
}

func (c *Container) RegisterBeanDef(name string, def *BeanDef) {
	if name == "" {
		name = strutil.FirstToLower(def.OriginType.Name())
	}
	c.mu.Lock()
	defer c.mu.Unlock()
	c.beans[name] = def
	c.typeMapping[def.OriginType] = []string{name}
}

// LookupType checks if a type is registered in the container.
func (c *Container) LookupType(typ reflect.Type) bool {
	if typ == nil {
		return false
	}
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	c.mu.RLock()
	defer c.mu.RUnlock()
	_, exists := c.typeMapping[typ]
	return exists
}

// GetBean gets bean instance by bean name
func (c *Container) GetBean(name string) (any, bool) {
	if name == "" {
		return nil, false
	}

	c.mu.RLock()

	def, exists := c.beans[name]
	if !exists {
		c.mu.RUnlock()
		return nil, false
	}
	c.mu.RUnlock()

	if def.IsPrototype {
		return createPrototypeBean(c, def), true
	}

	return def.Value, true
}

// GetBeanByType gets bean instance by type
func (c *Container) GetBeanByType(typ reflect.Type) (any, bool) {
	if typ == nil {
		return nil, false
	}
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	c.mu.RLock()
	names, exists := c.typeMapping[typ]
	c.mu.RUnlock()

	if !exists || len(names) == 0 {
		return nil, false
	}

	if len(names) == 1 {
		return c.GetBean(names[0])
	}

	panic("multiple beans found for type: " + typ.String())
}

// GetAllBeansByType retrieves all beans that implement the specified type
func (c *Container) GetAllBeansByType(typ reflect.Type) ([]any, bool) {
	if typ == nil {
		return nil, false
	}
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	c.mu.RLock()
	names, exists := c.typeMapping[typ]
	if !exists || len(names) == 0 {
		c.mu.RUnlock()
		return nil, false
	}
	c.mu.RUnlock()

	beans := make([]any, 0, len(names))
	for _, name := range names {
		if def, exists := c.beans[name]; exists {
			if def.IsPrototype {
				prototypeBean := createPrototypeBean(c, def)
				beans = append(beans, prototypeBean)
				continue
			}
			beans = append(beans, def.Value)
		}
	}

	return beans, len(beans) > 0
}

// RegisterBean registers an already instantiated bean instance into the IOC container.
// This method is used to register fully initialized bean instances that do not require
// dependency injector or lifecycle management by the container.
//
// Parameters:
//   - name: the unique bean name for registration
//   - instance: a pointer to the struct instance that has already been instantiated
//   - itypes: optional interface types that the instance implements, used for type-based lookup
//
// Notes:
//   - All beans registered by this method are treated as singletons.
//   - If a bean with the specified name already exists, an error will be returned.
//   - During registration, type-to-bean-name mappings are automatically established to support type-based bean retrieval.
//   - Unlike PreRegisterBean, this method accepts any struct pointer and does not require embedding Bean or mvc.Controller.
//   - The registered bean instances are immediately available for use and will not go through the container's lifecycle management.
func (c *Container) RegisterBean(name string, instance any, itypes ...reflect.Type) {
	if name == "" || instance == nil {
		panic("bean name and instance cannot be empty")
	}

	beanTyp := reflect.TypeOf(instance)
	if beanTyp.Kind() != reflect.Ptr || beanTyp.Elem().Kind() != reflect.Struct {
		panic("instance must be a pointer to a struct")
	}

	structType := beanTyp.Elem()

	c.mu.RLock()
	if _, exists := c.beans[name]; exists {
		c.mu.RUnlock()
		panic(fmt.Sprintf("bean with name '%s' already exists", name))
	}
	c.mu.RUnlock()

	c.mu.Lock()
	defer c.mu.Unlock()

	for _, itype := range itypes {
		if !beanTyp.Implements(itype) {
			continue
		}
		c.typeMapping[itype] = append(c.typeMapping[itype], name)
	}

	if _, exists := c.beans[name]; !exists {
		c.beans[name] = &BeanDef{
			ready: true,
			Value: instance,
		}
		c.typeMapping[structType] = append(c.typeMapping[structType], name)

		if _, ok := instance.(mvc.AbstractController); ok {
			ctrlType := reflect.TypeOf((*mvc.AbstractController)(nil)).Elem()
			c.typeMapping[ctrlType] = append(c.typeMapping[ctrlType], name)
		}
	}
}

// Refresh refreshes the container by processing all bean definitions.
// It creates beans in the order defined by their dependencies and clears the vars after creation.
// This method should be called only once, typically during application startup.
// It ensures that all beans are created and dependencies are injected correctly.
// If any bean creation fails, it logs a fatal error and stops the application.
func (c *Container) Refresh() {
	c.once.Do(func() {
		ctrlType := reflect.TypeOf((*mvc.AbstractController)(nil)).Elem()

		for beanName, beanDef := range c.beans {
			if beanDef.ready {
				continue
			}

			analyzeDefinition(c, beanDef)

			err := doProcessFields(c, reflect.ValueOf(beanDef.Value).Elem(), beanDef.AutowireFields)
			if err != nil {
				panic(fmt.Sprintf("failed to initialize bean '%s': %s", beanName, err.Error()))
			}
		}

		initializeBeans(c, ctrlType)
		injector.CleanWireConfigCache()
	})
}
