package container

import (
	"fmt"
	"sync"
)

// BeanDefinitionRegistry stores candidate bean definitions before the runtime
// container is refreshed. It keeps registration side effects separate from the
// effective bean graph built after configuration is available.
type BeanDefinitionRegistry struct {
	mu          sync.Mutex
	definitions map[string]*BeanDef
}

// NewBeanDefinitionRegistry initializes a registry for candidate bean definitions.
func NewBeanDefinitionRegistry() *BeanDefinitionRegistry {
	return &BeanDefinitionRegistry{
		definitions: make(map[string]*BeanDef),
	}
}

// RegisterBeanDef registers a candidate bean definition.
func (r *BeanDefinitionRegistry) RegisterBeanDef(name string, def *BeanDef) {
	if def == nil {
		panic("[BeanRegistry] registration failed: bean definition cannot be empty")
	}
	if name == "" {
		name = def.OriginType.String()
	}

	r.mu.Lock()
	defer r.mu.Unlock()

	if old, exists := r.definitions[name]; exists {
		panic(fmt.Sprintf("[BeanRegistry] bean name conflict: '%s' is already registered as %v, cannot register %v",
			name, old.Type, def.Type))
	}

	r.definitions[name] = def
}

// DrainDefs returns all currently registered bean definitions
func (r *BeanDefinitionRegistry) DrainDefs() map[string]*BeanDef {
	return r.definitions
}
