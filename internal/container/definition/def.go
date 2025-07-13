package definition

import (
	"reflect"
)

const (
	AutowireTag = "autowire" // tag for autowiring fields
	ConfTag     = "conf"     // tag for configuration fields
)

// AutowireField represents a field that requires autowiring
type AutowireField struct {
	Index       int                 // field index in struct
	IsInterface bool                // whether the field is an interface type
	Name        string              // field name
	AutowireTag string              // autowire tag value
	ConfTag     string              // configuration tag value
	Field       reflect.StructField // field reflection info

}

// BeanDefinition represents bean definition information
type BeanDefinition struct {
	IsPrototype    bool             // whether the bean is a prototype
	Name           string           // name of the bean
	Bean           any              // actual bean instance
	PtrType        reflect.Type     // point type of the bean
	OriginType     reflect.Type     // origin type of the bean
	AutowireFields []*AutowireField // fields that require autowiring
}

var (
	idx    = 0
	lookup = make(map[reflect.Type]struct{})
	defs   = make([]*BeanDefinition, 0, 8)
)

func LookupType(originType reflect.Type) bool {
	_, exists := lookup[originType]
	return exists
}

func RegisterBeanDefinition(def *BeanDefinition) {
	defs = append(defs, def)
	lookup[def.OriginType] = struct{}{}
}

func HasNext() bool {
	return idx < len(defs)
}

func Pop() *BeanDefinition {
	def := defs[idx]
	idx++

	return def
}

func GetAllDefinitions() []*BeanDefinition {
	return defs
}

func Clean() {
	defs = nil
	lookup = nil
}
