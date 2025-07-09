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
	// BeanName returns the name of the bean.
	// If the name is empty, it defaults to the struct name with the first letter in lowercase.
	BeanName() string

	// IsPrototype indicates whether the bean is a prototype.
	IsPrototype() bool
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
//	 func (s *UserService) BeanPostConstruct(beanName string) {
//	     // initialization logic here
//	 }
type BeanPostConstruct interface {
	// BeanPostConstruct is invoked after the bean is instantiated
	// and all its dependencies have been injected.
	BeanPostConstruct(beanName string)
}
