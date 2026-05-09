package bean

import "github.com/archine/gin-plus/v4/component/config"

// AbstractBean defines the basic contract for beans managed by the dependency injector container.
// Any struct implementing this interface can be recognized and managed as a bean.
type AbstractBean interface {
	// BeanName returns the unique name of the bean.
	// If the returned name is empty, the container will use package.structName as the default.
	// This name is used for bean identification and dependency injector.
	BeanName() string

	// Condition decides whether this bean should be registered.
	// The container calls Condition at registration time, passing a config.Provider.
	// Return true to register the bean; return false to skip registration.
	// The default implementation should return true.
	Condition(cfg config.Provider) bool
}

// Bean is a base struct that can be embedded into other structs to mark them as beans.
// Embedding Bean provides default implementations for AbstractBean methods.
//
// Usage:
//
//	type UserService struct {
//	    bean.Bean
//	    // your fields...
//	}
type Bean struct{}

func (b *Bean) BeanName() string {
	return ""
}

func (b *Bean) Condition(cp config.Provider) bool {
	return true
}

// PostConstruct defines a lifecycle callback interface for bean initialization.
// Any bean that implements this interface will have its BeanPostConstruct method
// automatically invoked by the container after instantiation and dependency injector.
//
// Usage:
//
//	type UserService struct {
//	    // your fields...
//	}
//
//	func (s *UserService) BeanPostConstruct() {
//	    // initialization logic here
//	}
type PostConstruct interface {
	// BeanPostConstruct is invoked after the bean is instantiated
	// and all its dependencies have been injected.
	BeanPostConstruct()
}
