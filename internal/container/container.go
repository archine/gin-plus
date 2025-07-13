package container

import (
	"errors"
	"fmt"
	"github.com/archine/gin-plus/v4/component/gplog"
	"github.com/archine/gin-plus/v4/component/mvc"
	"github.com/archine/gin-plus/v4/internal/container/definition"
	"reflect"
)

// BeanDef represents a bean definition in the IOC container.
type BeanDef struct {
	// ready indicates whether the bean is ready for use.
	ready bool

	// Value is the actual bean instance.
	value any

	// Typ is the origin type of the bean instance.
	originTyp reflect.Type

	// IsPrototype indicates whether the bean is a prototype (new instance for each request)
	isPrototype bool

	// AutowireFields contains fields that need to be autowired.
	autowireFields []*definition.AutowireField
}

// Container responsible for managing the lifecycle of all beans
type Container struct {
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
		return createPrototypeBean(c, def), true
	}

	return def.value, true
}

// GetBeanByType gets bean instance by type
func (c *Container) GetBeanByType(typ reflect.Type) (any, bool) {
	if typ == nil {
		return nil, false
	}
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	names, exists := c.typeMapping[typ]

	if !exists || len(names) == 0 {
		return nil, false
	}

	if len(names) == 1 {
		return c.GetBean(names[0])
	}

	gplog.Fatal(fmt.Sprintf("Multiple beans found for type '%s'. Please specify a bean name.", typ.String()))
	return nil, false
}

// GetAllBeansByType retrieves all beans that implement the specified type
func (c *Container) GetAllBeansByType(typ reflect.Type) ([]any, bool) {
	if typ == nil {
		return nil, false
	}
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	names, exists := c.typeMapping[typ]
	if !exists || len(names) == 0 {
		return nil, false
	}

	beans := make([]any, 0, len(names))
	for _, name := range names {
		if def, exists := c.beans[name]; exists {
			if def.isPrototype {
				prototypeBean := createPrototypeBean(c, def)
				beans = append(beans, prototypeBean)
				continue
			}
			beans = append(beans, def.value)
		}
	}

	return beans, len(beans) > 0
}

// RegisterBean registers an already instantiated bean instance into the IOC container.
//
// Parameters:
//   - name: the bean name for registration.
//   - instance: a pointer to the struct instance that has already been instantiated.
//   - itypes: optional interface types that the object implements, used for type-based lookup.
//
// Note:
//   - According to the IOC container design, all beans registered by this method are singletons.
//   - If a bean with the specified name already exists, an error will be returned.
//   - During registration, type-to-bean-name mappings are automatically established to support type-based bean retrieval.
func (c *Container) RegisterBean(name string, instance any, itypes ...reflect.Type) error {
	if name == "" || instance == nil {
		return errors.New("bean name and instance must not be empty or nil")
	}

	beanTyp := reflect.TypeOf(instance)
	if beanTyp.Kind() != reflect.Ptr || beanTyp.Elem().Kind() != reflect.Struct {
		return errors.New("instance must be a pointer to a struct")
	}

	structType := beanTyp.Elem()

	if _, exists := c.beans[name]; exists {
		return fmt.Errorf("duplicate bean name '%s'", name)
	}

	for _, itype := range itypes {
		if !beanTyp.Implements(itype) {
			return fmt.Errorf("instance type '%s' does not implement interface '%s'", beanTyp.Name(), itype.Name())
		}
		c.typeMapping[itype] = append(c.typeMapping[itype], name)
	}

	c.beans[name] = &BeanDef{
		ready: true,
		value: instance,
	}
	c.typeMapping[structType] = append(c.typeMapping[structType], name)

	if _, ok := instance.(mvc.AbstractController); ok {
		c.typeMapping[ctrlType] = append(c.typeMapping[ctrlType], name)
	}
	return nil
}
