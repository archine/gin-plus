package container

import (
	"errors"
	"fmt"
	"github.com/archine/gin-plus/v4/component/ioc"
	"github.com/archine/gin-plus/v4/internal/container/injection"
	"reflect"
	"sync"

	"github.com/archine/gin-plus/v4/component/log"
	"github.com/archine/gin-plus/v4/component/mvc"
	"github.com/archine/gin-plus/v4/internal/container/registry"
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
	autowireFields []*registry.AutowireField
}

// Container responsible for managing the lifecycle of all beans
type Container struct {
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

	log.Fatal(fmt.Sprintf("Multiple beans found for type '%s'. Please specify a bean name.", typ.String()))
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
// This method is used to register fully initialized bean instances that do not require
// dependency injection or lifecycle management by the container.
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
//
// Returns:
//   - error: nil on success, or an error if registration fails
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
		ctrlType := reflect.TypeOf((*mvc.AbstractController)(nil)).Elem()
		c.typeMapping[ctrlType] = append(c.typeMapping[ctrlType], name)
	}
	return nil
}

// Refresh refreshes the container by processing all bean definitions.
// It creates beans in the order defined by their dependencies and clears the vars after creation.
// This method should be called only once, typically during application startup.
// It ensures that all beans are created and dependencies are injected correctly.
// If any bean creation fails, it logs a fatal error and stops the application.
func (c *Container) Refresh() {
	c.once.Do(func() {
		err := processDefinitions(c)
		if err != nil {
			log.Fatal("Failed to process bean definitions: " + err.Error())
		}

		ctrlType := reflect.TypeOf((*mvc.AbstractController)(nil)).Elem()

		for beanName, bean := range c.beans {
			if bean.ready {
				continue
			}

			if err := inject(c, reflect.ValueOf(bean.value).Elem(), bean.autowireFields); err != nil {
				log.Fatal(fmt.Sprintf("Failed to inject dependencies for bean '%s': %s", beanName, err.Error()))
			}
		}

		refresh(c, ctrlType)
		injection.CleanInjectCache()
	})
}

// createPrototypeBean creates a prototype bean instance
func createPrototypeBean(c *Container, def *BeanDef) any {
	beanValue := reflect.New(def.originTyp)

	err := inject(c, beanValue.Elem(), def.autowireFields)
	if err != nil {
		log.Fatal(fmt.Sprintf("Failed to create prototype bean '%s': %s", def.originTyp.Name(), err.Error()))
	}

	beanValueIf := beanValue.Interface()

	if postConstruct, ok := beanValueIf.(ioc.BeanPostConstruct); ok {
		postConstruct.BeanPostConstruct()
	}

	return beanValueIf
}

// injectDependency injects a single dependency into a bean field
func inject(c *Container, structValue reflect.Value, autoFields []*registry.AutowireField) error {
	for _, autoField := range autoFields {

		if autoField.ValueTag != "" {
			injection.InjectConfig(structValue.Field(autoField.Index), autoField.Field.Type, autoField.ValueTag)
			continue
		}

		var bean any
		var exist bool

		if autoField.AutowireTag == "-" {
			bean, exist = c.GetBeanByType(autoField.Field.Type)
		} else {
			bean, exist = c.GetBean(autoField.AutowireTag)
		}

		if !exist {
			return fmt.Errorf("not found bean for field '%s' in '%s'", autoField.Name, structValue.Type().String())
		}

		err := injection.InjectBean(bean, structValue.Field(autoField.Index), autoField)
		if err != nil {
			return err
		}
	}

	return nil
}

func refresh(c *Container, ctrlType reflect.Type) {
	for beanName, bean := range c.beans {
		if bean.ready {
			continue
		}

		if postConstruct, ok := bean.value.(ioc.BeanPostConstruct); ok {
			postConstruct.BeanPostConstruct()
		}

		if _, ok := bean.value.(mvc.AbstractController); ok {
			c.typeMapping[ctrlType] = append(c.typeMapping[ctrlType], beanName)
		}

		bean.ready = true
	}
}
