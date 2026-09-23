package container

import (
	"errors"
	"fmt"
	"reflect"
	"sync"

	"github.com/archine/gin-plus/v4/component/config"
)

var (
	// ErrBeanNotFound indicates that no bean matches the requested name or type.
	ErrBeanNotFound = errors.New("no matching bean found")

	// ErrAmbiguousBean indicates that more than one bean matches a type-based lookup.
	ErrAmbiguousBean = errors.New("ambiguous bean dependency")
)

// BeanDef represents the definition and lifecycle state of a bean.
type BeanDef struct {
	Name       string                     // Unique bean name
	Condition  func(config.Provider) bool // Optional refresh-time registration guard
	Value      any                        // Actual instance (always a non-nil struct pointer)
	Type       reflect.Type               // Full type information (e.g., *UserService)
	OriginType reflect.Type               // The underlying struct type (e.g., UserService)

	deps  []*BeanDef // Beans injected into this bean, used to order PostConstruct callbacks
	ready bool       // True once the bean is fully initialized (PostConstruct done or not applicable)
}

// NewBeanDef validates instance and builds its definition.
// instance must be a non-nil struct pointer; an empty name falls back to the
// qualified struct name, e.g. "service.UserService".
func NewBeanDef(name string, instance any) (*BeanDef, error) {
	typ := reflect.TypeOf(instance)
	if typ == nil || typ.Kind() != reflect.Pointer || typ.Elem().Kind() != reflect.Struct {
		return nil, fmt.Errorf("expected a struct pointer, got %T", instance)
	}
	if reflect.ValueOf(instance).IsNil() {
		return nil, fmt.Errorf("instance of %v must not be a nil pointer", typ)
	}
	if name == "" {
		name = typ.Elem().String()
	}
	return &BeanDef{Name: name, Value: instance, Type: typ, OriginType: typ.Elem()}, nil
}

// assignableTo reports whether the bean can be injected where typ is expected.
// typ is either an interface type or a (pointer-normalized) struct type.
func (d *BeanDef) assignableTo(typ reflect.Type) bool {
	if typ.Kind() == reflect.Interface {
		return d.Type.Implements(typ)
	}
	return d.OriginType == typ
}

// Container manages the registration, dependency injection, and lifecycle of beans.
//
// Concurrency: all state is guarded by mu. Lookups are read-mostly, so type resolutions
// are cached as immutable snapshots (copy-on-write): readers share a snapshot without
// copying, and a registration never mutates one but drops the whole cache instead.
type Container struct {
	mu        sync.RWMutex
	refreshed bool
	beans     map[string]*BeanDef
	order     []*BeanDef                  // Registration order; drives every iteration for determinism
	typeCache map[reflect.Type][]*BeanDef // Resolved type → read-only snapshot (negative results included)
}

// NewContainer initializes a new IOC container.
func NewContainer() *Container {
	return &Container{
		beans:     make(map[string]*BeanDef),
		typeCache: make(map[reflect.Type][]*BeanDef),
	}
}

// GetBean retrieves a bean instance by its registered name.
func (c *Container) GetBean(name string) (any, bool) {
	c.mu.RLock()
	defer c.mu.RUnlock()

	if def, ok := c.beans[name]; ok {
		return def.Value, true
	}
	return nil, false
}

// GetBeanByType retrieves the single bean matching typ (struct pointer or interface).
// It returns ErrBeanNotFound when nothing matches and ErrAmbiguousBean when several beans match.
func (c *Container) GetBeanByType(typ reflect.Type) (any, error) {
	def, err := c.resolveUnique(typ)
	if err != nil {
		return nil, err
	}
	return def.Value, nil
}

// GetAllBeansByType retrieves all bean instances matching typ, in registration order.
func (c *Container) GetAllBeansByType(typ reflect.Type) []any {
	defs := c.lookup(typ)
	if len(defs) == 0 {
		return nil
	}

	values := make([]any, len(defs))
	for i, def := range defs {
		values[i] = def.Value
	}
	return values
}

// RegisterBean registers a pre-instantiated object as a ready singleton bean.
// Such beans skip dependency injection and PostConstruct; they are resolvable by name,
// by their struct type and by every interface they implement.
// If name is empty, the qualified struct name is used. Panics on invalid input or name conflict.
func (c *Container) RegisterBean(name string, instance any) {
	def, err := NewBeanDef(name, instance)
	if err != nil {
		panic("[IOC] registration failed: " + err.Error())
	}
	def.ready = true

	c.mu.Lock()
	defer c.mu.Unlock()
	if err = c.registerLocked(def); err != nil {
		panic(err.Error())
	}
}

// Refresh builds the effective bean graph from the candidate definitions:
//  1. evaluates each Condition and registers the accepted definitions;
//  2. injects config values and dependencies into every registered definition;
//  3. runs PostConstruct callbacks so that dependencies are initialized first.
//
// It must be invoked exactly once, after configuration is loaded.
func (c *Container) Refresh(cp config.Provider, defs []*BeanDef) error {
	if cp == nil {
		return errors.New("[IOC] config provider is not initialized before container refresh")
	}

	// Conditions are user code and may query the container, so they must run without holding the lock.
	active := make([]*BeanDef, 0, len(defs))
	for _, def := range defs {
		if def.Condition == nil || def.Condition(cp) {
			active = append(active, def)
		}
	}

	c.mu.Lock()
	if c.refreshed {
		c.mu.Unlock()
		return errors.New("[IOC] container has already been refreshed")
	}
	c.refreshed = true
	for _, def := range active {
		if err := c.registerLocked(def); err != nil {
			c.mu.Unlock()
			return err
		}
	}
	c.mu.Unlock()

	// All candidates are registered before injection starts, so declaration order never matters.
	for _, def := range active {
		if err := c.inject(cp, def); err != nil {
			return fmt.Errorf("[IOC] failed to initialize bean '%s' (%v): %w", def.Name, def.Type, err)
		}
	}

	postConstruct(active)
	return nil
}

// registerLocked adds def to the container. Caller must hold the write lock.
func (c *Container) registerLocked(def *BeanDef) error {
	if old, exists := c.beans[def.Name]; exists {
		return fmt.Errorf("[IOC] bean name conflict: '%s' is already registered as %v, cannot register %v",
			def.Name, old.Type, def.Type)
	}
	c.beans[def.Name] = def
	c.order = append(c.order, def)
	clear(c.typeCache) // Drop snapshots rather than patch them: any of them may now be incomplete
	return nil
}

// resolveUnique returns the only bean assignable to typ.
func (c *Container) resolveUnique(typ reflect.Type) (*BeanDef, error) {
	defs := c.lookup(typ)
	switch len(defs) {
	case 0:
		return nil, fmt.Errorf("%w: type %v", ErrBeanNotFound, typ)
	case 1:
		return defs[0], nil
	default:
		names := make([]string, len(defs))
		for i, def := range defs {
			names[i] = def.Name
		}
		return nil, fmt.Errorf("%w: type %v matches beans %v", ErrAmbiguousBean, typ, names)
	}
}

// lookup returns the definitions assignable to typ in registration order.
// Pointer types are normalized to their element type, so *UserService and UserService
// resolve identically.
//
// The returned slice is a shared snapshot and must be treated as read-only.
func (c *Container) lookup(typ reflect.Type) []*BeanDef {
	if typ == nil {
		return nil
	}
	if typ.Kind() == reflect.Pointer {
		typ = typ.Elem()
	}

	// Fast path: return the cached snapshot as-is, no allocation.
	c.mu.RLock()
	defs, cached := c.typeCache[typ]
	c.mu.RUnlock()
	if cached {
		return defs
	}

	// Slow path: build the snapshot under the write lock. Double-check first, since
	// another goroutine may have populated the entry in between.
	c.mu.Lock()
	defer c.mu.Unlock()
	if defs, cached = c.typeCache[typ]; cached {
		return defs
	}
	for _, def := range c.order {
		if def.assignableTo(typ) {
			defs = append(defs, def)
		}
	}
	// Cap to length so that an accidental append by a reader reallocates instead of
	// writing into the shared snapshot. A nil entry caches the negative result.
	defs = defs[:len(defs):len(defs)]
	c.typeCache[typ] = defs
	return defs
}
