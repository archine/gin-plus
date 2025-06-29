package ctrl

import "github.com/archine/gin-plus/v4/component/ioc"

// Controller is a base interface that declares an entity as a controller.
// When the application starts, it will automatically scan for structs that implement this interface,
// register them as beans in the IOC container, and register their methods as API endpoints.
//
// Usage:
//
//	type UserController struct {
//	    ctrl.Controller
//	}
//
//	@GET(path="/user/:id")
//	func (c *UserController) GetUser(ctx *gin.Context) {
//	    // Handle user retrieval logic
//	}
type Controller interface {
	ioc.Bean
}
