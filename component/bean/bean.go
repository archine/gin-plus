package bean

// AbstractBean is the interface that all beans must implement.
// It provides methods for bean lifecycle management and naming.
// Beans are objects managed by the IoC container, and they can be automatically instantiated
// and injected into other components as needed.
// Beans can be used to encapsulate business logic, manage dependencies, and provide a consistent way
// to access shared resources within the application.
// If you don't want to implement all the methods, you can rewrite some of them by combining beans in the structure.
type AbstractBean interface {
	// BeanPostConstruct is called after the bean is instantiated and its properties are set.
	BeanPostConstruct()

	// BeanName returns the name of the bean.
	// If it is not set, the structure name with the first letter in lowercase is used.
	BeanName() string
}

// Bean declare this structure as a Bean.
// When other beans inject this bean, the structure will be automatically instantiated
// and its properties processed when it does not exist in the container
type Bean struct{}

func (b *Bean) BeanPostConstruct() {}

func (b *Bean) BeanName() string {
	return ""
}
