package registry

import (
	"reflect"
	"sync"
)

// AutowireField represents a field that requires autowiring
type AutowireField struct {
	Index       int                 // field index in struct
	IsInterface bool                // whether the field is an interface type
	Name        string              // field name
	AutowireTag string              // autowire tag value
	ValueTag    string              // value tag value
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

// BeanRegistry manages bean definitions
type BeanRegistry struct {
	mu     sync.RWMutex
	idx    int
	lookup map[reflect.Type]struct{}
	defs   []*BeanDefinition
}

var defaultRegistry = NewBeanRegistry()

// NewBeanRegistry creates a new bean registry
func NewBeanRegistry() *BeanRegistry {
	return &BeanRegistry{
		lookup: make(map[reflect.Type]struct{}),
		defs:   make([]*BeanDefinition, 0),
	}
}

// IsTypeRegistered checks if a type is already registered
func (r *BeanRegistry) IsTypeRegistered(originType reflect.Type) bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	_, exists := r.lookup[originType]
	return exists
}

// Register registers a bean definition
func (r *BeanRegistry) Register(def *BeanDefinition) {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.defs = append(r.defs, def)
	r.lookup[def.OriginType] = struct{}{}
}

// HasNext checks if there are more definitions to process
func (r *BeanRegistry) HasNext() bool {
	r.mu.RLock()
	defer r.mu.RUnlock()
	return r.idx < len(r.defs)
}

// Next returns the next bean definition
func (r *BeanRegistry) Next() *BeanDefinition {
	r.mu.Lock()
	defer r.mu.Unlock()
	if r.idx >= len(r.defs) {
		return nil
	}
	def := r.defs[r.idx]
	r.idx++
	return def
}

// GetAll returns all registered bean definitions
func (r *BeanRegistry) GetAll() []*BeanDefinition {
	r.mu.RLock()
	defer r.mu.RUnlock()
	result := make([]*BeanDefinition, len(r.defs))
	copy(result, r.defs)
	return result
}

// Reset resets the registry index
func (r *BeanRegistry) Reset() {
	r.mu.Lock()
	defer r.mu.Unlock()
	r.idx = 0
}

func LookupType(typ reflect.Type) bool {
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}
	return defaultRegistry.IsTypeRegistered(typ)
}

func RegisterBeanDefinition(def *BeanDefinition) {
	defaultRegistry.Register(def)
}

func HasNext() bool {
	return defaultRegistry.HasNext()
}

func Pop() *BeanDefinition {
	return defaultRegistry.Next()
}

func GetAllDefinitions() []*BeanDefinition {
	return defaultRegistry.GetAll()
}

func Clean() {
	defaultRegistry = nil
}
