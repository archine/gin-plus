package definition

import (
	"reflect"
)

// AutowireField represents a field that requires autowiring
type AutowireField struct {
	Index       int                 // field index in struct
	Name        string              // field name
	IsInterface bool                // whether field is interface type
	AutowireTag string              // autowire tag value
	Field       reflect.StructField // field reflection info
}

// BeanDefinition represents bean definition information
type BeanDefinition struct {
	IsPrototype    bool             // whether the bean is a prototype
	Name           string           // name of the bean, defaults to struct name with first letter lowercase
	Bean           any              // actual bean instance, if available
	Type           reflect.Type     // type of the bean
	AutowireFields []*AutowireField // fields that require autowiring
}

// Cache holds the registered bean definitions
var Cache = make(map[string]*BeanDefinition)
