package ioc

import (
	"fmt"
	"reflect"

	"github.com/archine/gin-plus/v4/component/gplog"
	"github.com/archine/gin-plus/v4/component/ioc/internal/topo"
)

// refresh initializes the container and creates all registered beans
//
// Note: this function is system-internal and should not be used directly in application code.
func refresh() {
	defaultContainer.refresh()
}

func (c *Container) refresh() {
	once.Do(func() {
		refreshed.Store(true)

		creationOrder, err := c.buildCreationOrder()
		if err != nil {
			gplog.Fatal("Failed to build bean creation order: " + err.Error())
		}

		for _, beanName := range creationOrder {
			if err = c.createBean(beanName); err != nil {
				gplog.Fatal(fmt.Sprintf("Failed to create bean '%s': %s", beanName, err.Error()))
			}
		}

		c.definitions = nil // clear definitions after creation
		beanType = nil      // clear bean type to avoid memory leaks
		lazyBeanType = nil  // clear lazy bean type to avoid memory leaks
	})
}

// buildCreationOrder builds bean creation order using topological sorting
func (c *Container) buildCreationOrder() ([]string, error) {
	var edges []*topo.DependencyEdge

	for beanName, def := range c.definitions {
		for _, dependency := range def.DependentBeans {
			edge := &topo.DependencyEdge{
				From: beanName,
				To:   dependency,
			}
			edges = append(edges, edge)
		}
	}

	if len(edges) == 0 {
		return nil, nil
	}

	return topo.Sort(edges)
}

// createBean creates a bean instance and performs dependency injection
func (c *Container) createBean(beanName string) error {
	c.mu.Lock()
	defer c.mu.Unlock()

	if _, exists := c.beans[beanName]; exists {
		return nil // already created
	}

	def, exists := c.definitions[beanName]
	if !exists {
		return fmt.Errorf("bean definition not found for '%s'", beanName)
	}

	beanValue := reflect.New(def.Type).Elem()

	// inject dependencies
	for _, depField := range def.DependentFields {
		if err := c.injectDependency(beanValue, depField); err != nil {
			return fmt.Errorf("failed to inject dependency for field '%s': %w",
				depField.Field.Name, err)
		}
	}

	beanInstance := beanValue.Interface()

	// call post-construct if available
	if postConstruct, ok := beanInstance.(BeanPostConstruct); ok {
		postConstruct.BeanPostConstruct()
	}

	c.beans[beanName] = beanInstance
	return nil
}

// injectDependency injects a single dependency into a bean field
func (c *Container) injectDependency(beanValue reflect.Value, depField *DependencyField) error {
	if depField.AutowireTag == "-" {
		return nil
	}

	dependencyBean, exists := c.beans[depField.AutowireTag]
	if !exists {
		return fmt.Errorf("dependency bean '%s' not found", depField.AutowireTag)
	}

	dependencyType := reflect.TypeOf(dependencyBean)
	fieldValue := beanValue.Field(depField.Index)

	if depField.IsInterface {
		if !dependencyType.Implements(depField.Field.Type) {
			return fmt.Errorf("bean '%s' does not implement interface '%s'",
				depField.AutowireTag, depField.Field.Type.String())
		}
		fieldValue.Set(reflect.ValueOf(dependencyBean))
		return nil
	}

	if !dependencyType.AssignableTo(depField.Field.Type) {
		return fmt.Errorf("type mismatch: expected '%s', got '%s'",
			depField.Field.Type.String(), dependencyType.String())
	}

	fieldValue.Set(reflect.ValueOf(dependencyBean))
	return nil
}
