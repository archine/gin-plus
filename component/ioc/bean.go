package ioc

// Bean marks a struct as a Bean in the dependency injection container.
// Any struct implementing this interface will be recognized and managed as a Bean.
//
// Usage:
//
//	type UserService struct {
//	    ioc.Bean
//	    // your fields...
//	}
type Bean interface {
	// IsBean is a marker method with no implementation required.
	IsBean()
}

// LazyBean defines the interface for lazy-loaded Bean components.
// Beans implementing this interface can control their instantiation timing.
//
// Usage:
//
//	type UserService struct {
//	    ioc.LazyBean
//	    // your fields...
//	}
type LazyBean interface {
	Bean
	// IsLazyBean is a marker method with no implementation required.
	IsLazyBean()
}

// BeanPostConstruct defines the lifecycle callback interface for Bean initialization.
// Beans implementing this interface will have their BeanPostConstruct method
// called automatically after instantiation and dependency injection.
//
// Usage:
//
//		type UserService struct {
//		    // your fields...
//		}
//
//	 func (s *UserService) BeanPostConstruct() {
//	     // initialization logic here
//	 }
type BeanPostConstruct interface {
	// BeanPostConstruct is invoked after the bean is instantiated
	// and all its dependencies have been injected.
	BeanPostConstruct()
}
