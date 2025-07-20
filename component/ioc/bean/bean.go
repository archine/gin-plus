package bean

// AbstractBean defines the basic contract for beans managed by the dependency injector container.
// Any struct implementing this interface can be recognized and managed as a bean.
type AbstractBean interface {
	// BeanName returns the unique name of the bean.
	// If the returned name is empty, the container will use the struct name with the first letter in lowercase as the default.
	// This name is used for bean identification and dependency injector.
	BeanName() string

	// IsPrototype indicates whether the bean should be treated as a prototype.
	// If true, the container will create a new instance each time the bean is requested.
	// If false, the same singleton instance will be returned for every request.
	IsPrototype() bool
}

// Bean is a base struct that can be embedded into other structs to mark them as beans.
// Embedding Bean provides default implementations for AbstractBean methods.
//
// Usage:
//
//	type UserService struct {
//	    ioc.Bean
//	    // your fields...
//	}
type Bean struct{}

func (b *Bean) BeanName() string {
	return ""
}

func (b *Bean) IsPrototype() bool {
	return false
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
