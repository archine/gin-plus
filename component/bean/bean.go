package bean

// Marker marks a struct as a Bean component in the dependency injection container.
// Any struct implementing this interface will be recognized and managed as a Bean.
type Marker interface {
	// IsBean is a marker method with no implementation required.
	// It serves as a type constraint to identify Bean components.
	IsBean()
}

// PostConstruct defines the lifecycle callback interface for Bean initialization.
// Beans implementing this interface will have their BeanPostConstruct method
// called automatically after instantiation and dependency injection.
type PostConstruct interface {
	// BeanPostConstruct is invoked after the bean is instantiated and all its
	// dependencies have been injected. Use this method for initialization logic
	// that requires fully configured dependencies.
	BeanPostConstruct()
}

// Lazy defines the interface for lazy-loaded Bean components.
// Beans implementing this interface can control their instantiation timing.
type Lazy interface {
	// IsLazyBean indicates whether the bean should be instantiated lazily.
	// When true, the bean will only be created when first accessed or injected,
	// rather than during container initialization.
	IsLazyBean()
}
