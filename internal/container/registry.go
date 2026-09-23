package container

import (
	"fmt"
	"sync"
)

// BeanDefinitionRegistry stores candidate bean definitions before the runtime
// container is refreshed. It keeps registration side effects (typically from init
// functions) separate from the effective bean graph built after configuration is available.
//
// The registry is single-use: Drain hands the definitions over to the container and
// closes it, so late registrations fail loudly instead of being silently ignored.
type BeanDefinitionRegistry struct {
	mu     sync.Mutex
	defs   []*BeanDef          // Registration order, preserved into the container
	names  map[string]*BeanDef // Name index for conflict detection
	closed bool
}

// NewBeanDefinitionRegistry initializes a registry for candidate bean definitions.
func NewBeanDefinitionRegistry() *BeanDefinitionRegistry {
	return &BeanDefinitionRegistry{names: make(map[string]*BeanDef)}
}

// Register adds a candidate definition. It panics on a nil definition, a name conflict,
// or when the registry has already been drained by a container refresh.
func (r *BeanDefinitionRegistry) Register(def *BeanDef) {
	if def == nil {
		panic("[BeanRegistry] registration failed: bean definition cannot be nil")
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if r.closed {
		panic(fmt.Sprintf("[BeanRegistry] cannot register '%s': the container has already been refreshed, "+
			"use ioc.RegisterBean to register a ready instance instead", def.Name))
	}
	if old, exists := r.names[def.Name]; exists {
		panic(fmt.Sprintf("[BeanRegistry] bean name conflict: '%s' is already registered as %v, cannot register %v",
			def.Name, old.Type, def.Type))
	}

	r.names[def.Name] = def
	r.defs = append(r.defs, def)
}

// Drain returns all registered definitions in registration order and closes the registry.
func (r *BeanDefinitionRegistry) Drain() []*BeanDef {
	r.mu.Lock()
	defer r.mu.Unlock()

	defs := r.defs
	r.defs, r.names, r.closed = nil, nil, true
	return defs
}
