package event

// ContainerBeforeInitEvent handles events that occur before bean container initialization.
// Implement this interface to perform setup tasks before the container starts.
type ContainerBeforeInitEvent interface {
	// OnBeanContainerBeforeInit is called before the bean container begins initialization.
	// Use this to manually register beans or perform pre-initialization setup.
	// Note: No container beans are available at this stage.
	OnBeanContainerBeforeInit()
}

// ContainerAfterInitEvent handles events that occur after bean container initialization.
// Implement this interface to perform tasks that require access to initialized beans.
type ContainerAfterInitEvent interface {
	// OnBeanContainerAfterInit is called after the bean container finishes initialization.
	// Use this to configure beans, perform post-initialization tasks, or setup dependencies.
	// All registered beans are fully available and ready for use at this stage.
	OnBeanContainerAfterInit()
}
