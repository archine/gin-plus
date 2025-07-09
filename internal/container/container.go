package container

import (
	"errors"
	"fmt"
	"github.com/archine/gin-plus/v4/internal/container/definition"
	"reflect"
)

// BeanDef represents a bean definition in the IOC container.
type BeanDef struct {
	// Value is the actual bean instance.
	value any

	// Typ is the type of the bean instance.
	typ reflect.Type

	// IsPrototype indicates whether the bean is a prototype (new instance for each request)
	isPrototype bool

	// AutowireFields contains fields that need to be autowired.
	autowireFields []*definition.AutowireField
}

// Container responsible for managing the lifecycle of all beans
type Container struct {
	beans       map[string]*BeanDef       // stores bean definition information
	typeMapping map[reflect.Type][]string // stores type to bean names mapping
}

func NewContainer() *Container {
	return &Container{
		beans:       make(map[string]*BeanDef),
		typeMapping: make(map[reflect.Type][]string),
	}
}

// GetBean gets bean instance by bean name
func (c *Container) GetBean(name string) (any, bool) {
	if name == "" {
		return nil, false
	}

	def, exists := c.beans[name]
	if !exists {
		return nil, false
	}
	if def.isPrototype {
		// If it's a prototype bean, create a new instance
		prototypeBean, err := c.createPrototypeBean(def)
		if err != nil {
			return nil, false
		}
		return prototypeBean, true
	}
	return def.value, true
}

// GetBeanByType gets bean instance by type
func (c *Container) GetBeanByType(typ reflect.Type) (any, error) {
	if typ == nil {
		return nil, fmt.Errorf("type cannot be nil")
	}
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	names, exists := c.typeMapping[typ]

	if !exists || len(names) == 0 {
		return nil, fmt.Errorf("no beans found for type '%s'", typ.Name())
	}

	if len(names) == 1 {
		def, exists := c.beans[names[0]]
		if !exists {
			return nil, fmt.Errorf("bean '%s' not found in container", names[0])
		}
		if def.isPrototype {
			return c.createPrototypeBean(def)
		}
		return def.value, nil
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

	names, exists := c.typeMapping[typ]
	if !exists || len(names) == 0 {
		return nil, fmt.Errorf("no beans found for type '%s'", typ.Name())
	}

	beans := make([]any, 0, len(names))
	for _, name := range names {
		if def, exists := c.beans[name]; exists {
			if def.isPrototype {
				prototypeBean, err := c.createPrototypeBean(def)
				if err != nil {
					return nil, fmt.Errorf("failed to create prototype bean '%s': %w", name, err)
				}
				beans = append(beans, prototypeBean)
				continue
			}
			beans = append(beans, def.value)
		}
	}

	if len(beans) == 0 {
		return nil, fmt.Errorf("no beans found for type '%s'", typ.Name())
	}
	return beans, nil
}

// RegisterBean registers an already instantiated bean instance into the IOC container.
//
// Parameters:
//   - name: the bean name for registration.
//   - bean: a pointer to the struct instance that has already been instantiated.
//   - itypes: optional interface types that the object implements, used for type-based lookup.
//
// Note:
//   - According to the IOC container design, all beans registered by this method are singletons.
//   - If a bean with the specified name already exists, an error will be returned.
//   - During registration, type-to-bean-name mappings are automatically established to support type-based bean retrieval.
func (c *Container) RegisterBean(name string, bean any, itypes ...reflect.Type) error {
	if name == "" || bean == nil {
		return errors.New("bean name and bean instance must not be empty or nil")
	}

	beanTyp := reflect.TypeOf(bean)
	if beanTyp.Kind() != reflect.Ptr || beanTyp.Elem().Kind() != reflect.Struct {
		return errors.New("bean must be a pointer to a struct")
	}
	structType := beanTyp.Elem()

	if _, exists := c.beans[name]; exists {
		return fmt.Errorf("bean '%s' already exists", name)
	}

	for _, itype := range itypes {
		if !beanTyp.Implements(itype) {
			return fmt.Errorf("object type '%s' does not implement interface '%s'", beanTyp.Name(), itype.Name())
		}
		c.typeMapping[itype] = append(c.typeMapping[itype], name)
	}

	c.beans[name] = &BeanDef{
		value:       bean,
		isPrototype: false,
	}

	if _, exists := c.typeMapping[structType]; !exists {
		c.typeMapping[structType] = []string{name}
	}

	return nil
}
