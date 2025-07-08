package router

type Method struct {
	// NameOrFunc method name or gin.HandlerFunc,
	// the name is used to find the method in the controller, muse be first letter uppercase.
	NameOrFunc any

	// HttpMethod HTTP method (GET, POST, PUT, DELETE, PATCH, OPTIONS, HEAD)
	HttpMethod string

	// Path is the route path for the method.
	Path string
}

type Route struct {
	// Name is the bean name of the controller, used to find the controller in the IOC container.
	Name string

	// BasePath is the base path for the controller's routes.
	BasePath string

	// Methods is a list of methods associated with the controller.
	Methods []*Method
}
