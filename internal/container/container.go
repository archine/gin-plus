package container

import (
	"errors"
	"fmt"
	"reflect"

	"github.com/archine/gin-plus/v4/component/bean"
	"github.com/archine/gin-plus/v4/component/gplog"
	"github.com/archine/gin-plus/v4/exception"
)

var (
	componentType = reflect.TypeOf((*bean.Component)(nil)).Elem()
	lazeInitType  = reflect.TypeOf((*bean.LazyInit)(nil)).Elem()
)

// structField represents dependency injection information for struct fields
type structField struct {
	idx         int                 // field index position in struct
	name        string              // field name
	isInterface bool                // indicates if field is interface type
	autowire    string              // autowire tag value, used to specify bean name for injection
	field       reflect.StructField // reflection information of the field
}

// beanDefinition represents bean definition information, including type and dependency relationships
type beanDefinition struct {
	typ             reflect.Type   // bean type information
	dependFields    []*structField // list of fields that need dependency injection
	dependBeanNames []string       // list of dependent bean names
}

// BeanContainer IOC container, responsible for managing the lifecycle of all beans
type BeanContainer struct {
	refreshed    bool                       // indicates if container has completed initialization
	beans        map[string]any             // stores created bean instances, key is bean name (lowercase first letter)
	befCache     map[string]*beanDefinition // stores bean definition information, key is bean name
	typeForNames map[reflect.Type][]string  // stores mapping from type to bean names, used for type-based bean lookup
}

// container global container instance
var container = &BeanContainer{
	refreshed:    false,
	beans:        make(map[string]any),
	befCache:     make(map[string]*beanDefinition),
	typeForNames: make(map[reflect.Type][]string),
}

// GetBeanContainer gets the initialized bean container instance
//
// Returns:
//   - *BeanContainer: container instance if initialized, nil otherwise
func GetBeanContainer() *BeanContainer {
	if !container.refreshed {
		gplog.Warn("container has not been initialized")
		return nil
	}
	return container
}

// GetBean gets bean instance by bean name
//
// Args:
//   - name: bean name to search for
//
// Returns:
//   - any: bean instance if found
//   - bool: true if bean exists, false otherwise
func (bc *BeanContainer) GetBean(name string) (any, bool) {
	if name == "" {
		return nil, false
	}

	b, exists := bc.beans[name]
	return b, exists
}

// GetBeanByType gets bean instance by type
//
// Args:
//   - structValue: struct instance or pointer, used to determine the type to search for
//
// Returns:
//   - any: bean instance if found
//   - bool: true if bean exists, false otherwise
//
// Note: throws exception if multiple bean instances exist for the same type
func (bc *BeanContainer) GetBeanByType(structValue any) (any, bool) {
	if structValue == nil {
		return nil, false
	}

	typ := reflect.TypeOf(structValue)
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}

	names, ok := bc.typeForNames[typ]
	if !ok {
		return nil, false
	}

	if len(names) == 1 {
		b, exists := bc.GetBean(names[0])
		if exists {
			return b, true
		}
		return nil, false
	}

	// multiple bean instances exist for the same type
	panic(exception.NewStackErr(fmt.Sprintf("multiple beans found for type '%s': %v", typ.Name(), names)).ToString())
}

// RegisterBeanDefinition registers bean definitions to the container
//
// Args:
//   - structTypes: list of struct types to register that implement bean.Marker interface
//
// Returns:
//   - error: error information during registration process, nil if successful
func RegisterBeanDefinition(structTypes []reflect.Type) error {
	if container.refreshed {
		return errors.New("container already initialized")
	}

	if len(structTypes) == 0 {
		return nil
	}

	for _, typ := range structTypes {
		if typ == nil {
			continue
		}

		if typ.Kind() != reflect.Ptr && typ.Elem().Kind() != reflect.Struct {
			return fmt.Errorf("type '%s' must be struct pointer", typ.Name())
		}
		if typ.Implements(lazeInitType) {
			continue
		}
		if !typ.Implements(componentType) {
			return fmt.Errorf("type '%s' must embed bean.Component", typ.Name())
		}

		if typ.Kind() == reflect.Ptr {
			typ = typ.Elem()
		}

		beanName := typ.Name()

		// check if bean definition already exists
		if _, exists := container.befCache[beanName]; exists {
			return fmt.Errorf("bean '%s' already registered", beanName)
		}

		def := &beanDefinition{
			typ: typ,
		}

		// parse bean dependencies
		newStructTypes, err := getStructDependsBean(def)
		if err != nil {
			return fmt.Errorf("failed to parse bean '%s': %w", beanName, err)
		}

		container.befCache[beanName] = def
		container.typeForNames[typ] = []string{beanName}

		// recursively register dependent bean definitions
		if len(newStructTypes) > 0 {
			if err = RegisterBeanDefinition(newStructTypes); err != nil {
				gplog.Fatal(exception.NewStackErr("failed to register bean definitions: " + err.Error()).ToString())
			}
		}
	}

	return nil
}

// Refresh initializes the container and creates all registered bean instances
// This method determines bean creation order based on dependencies and completes dependency injection
func Refresh() {
	gplog.Info("starting bean container initialization...")

	// build bean creation order (topological sort)
	creationOrder, err := buildBeanCreatOrder()
	if err != nil {
		gplog.Fatal(exception.NewStackErr(fmt.Sprintf("failed to build creation order: %s", err.Error())).ToString())
	}

	for _, beanName := range creationOrder {
		if err = createBean(beanName); err != nil {
			gplog.Fatal(exception.NewStackErr(fmt.Sprintf("failed to create bean '%s': %s", beanName, err.Error())).ToString())
		}
	}

	container.befCache = nil
	componentType = nil
	lazeInitType = nil

	container.refreshed = true // mark container as initialized

	gplog.Info("bean container initialization completed")
}
