package event

// BeanPostProcessor defines the event interface for post-processing beans after their initialization.
// This interface allows custom logic to be executed after each bean has been created and configured
// in the dependency injection container.
type BeanPostProcessor interface {
	AppEvent

	// OnBeanPostProcess is called after a bean has been successfully initialized and configured.
	// This method receives the bean name and the actual bean instance, allowing for custom
	// post-processing operations such as validation, decoration, or additional configuration.
	//
	// Parameters:
	//   beanName - the name/identifier of the bean that was just initialized
	//   bean - the actual bean instance that was created and configured
	OnBeanPostProcess(beanName string, bean any)
}
