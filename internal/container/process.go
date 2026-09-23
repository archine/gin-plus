package container

import (
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/archine/gin-plus/v4/component/config"
	"github.com/archine/gin-plus/v4/component/ioc/bean"
	"github.com/archine/gin-plus/v4/internal/container/injector"
)

// inject populates every `value` and `autowire` field of def and records its bean dependencies.
// Embedded (anonymous) fields are skipped. A field carrying both tags is treated as a config value.
func (c *Container) inject(cp config.Provider, def *BeanDef) error {
	structValue := reflect.ValueOf(def.Value).Elem()

	for i := 0; i < def.OriginType.NumField(); i++ {
		field := def.OriginType.Field(i)
		if field.Anonymous {
			continue
		}

		if tag := field.Tag.Get(injector.ValueTag); tag != "" {
			expr, err := injector.ParseValueExpr(tag)
			if err == nil {
				err = injector.WireConfigValue(cp, structValue.Field(i), expr)
			}
			if err != nil {
				return fmt.Errorf("field '%s': %w", field.Name, err)
			}
			continue
		}

		tag, ok := field.Tag.Lookup(injector.AutowireTag)
		if !ok || tag == "-" {
			continue
		}

		dep, err := c.resolveDependency(field, tag)
		if err != nil {
			return fmt.Errorf("field '%s': %w", field.Name, err)
		}
		if dep == nil {
			continue // Optional dependency without a candidate: leave the zero value
		}

		injector.WireBean(dep.Value, structValue.Field(i))
		def.deps = append(def.deps, dep)
	}

	return nil
}

// resolveDependency finds the bean for an autowire field.
// Tag syntax: `autowire:""` (by type), `autowire:"name"` (by name), with an optional ",optional" suffix.
// It returns (nil, nil) only for an optional dependency that has no candidate.
// Ambiguity is always an error, even for optional fields, because silently skipping would hide a misconfiguration.
func (c *Container) resolveDependency(field reflect.StructField, tag string) (*BeanDef, error) {
	ft := field.Type
	if ft.Kind() != reflect.Interface && (ft.Kind() != reflect.Pointer || ft.Elem().Kind() != reflect.Struct) {
		return nil, fmt.Errorf("autowire target must be a struct pointer or an interface, got %v", ft)
	}

	name, option, _ := strings.Cut(tag, ",")
	name, option = strings.TrimSpace(name), strings.TrimSpace(option)
	if option != "" && option != "optional" {
		return nil, fmt.Errorf("unsupported autowire option '%s', only 'optional' is supported", option)
	}
	optional := option == "optional"

	if name == "" {
		dep, err := c.resolveUnique(ft)
		if optional && errors.Is(err, ErrBeanNotFound) {
			return nil, nil
		}
		return dep, err
	}

	c.mu.RLock()
	dep, exists := c.beans[name]
	c.mu.RUnlock()
	if !exists {
		if optional {
			return nil, nil
		}
		return nil, fmt.Errorf("%w: name '%s'", ErrBeanNotFound, name)
	}
	// Validate here so that a type mismatch surfaces as a clear error instead of a reflect panic.
	if !dep.Type.AssignableTo(ft) {
		return nil, fmt.Errorf("bean '%s' of type %v is not assignable to %v", name, dep.Type, ft)
	}
	return dep, nil
}

// postConstruct invokes BeanPostConstruct on every bean, dependencies first.
//
// A depth-first traversal over the recorded dependencies guarantees that a bean's
// dependencies are fully initialized before its own callback runs. Field injection
// permits cyclic dependencies; a cycle is broken at the bean currently being visited,
// so beans inside a cycle are initialized in registration order.
func postConstruct(defs []*BeanDef) {
	visiting := make(map[*BeanDef]bool, len(defs))

	var visit func(def *BeanDef)
	visit = func(def *BeanDef) {
		if def.ready || visiting[def] {
			return
		}
		visiting[def] = true

		for _, dep := range def.deps {
			visit(dep)
		}
		if pc, ok := def.Value.(bean.PostConstruct); ok {
			pc.BeanPostConstruct()
		}
		def.ready = true
	}

	for _, def := range defs {
		visit(def)
	}
}
