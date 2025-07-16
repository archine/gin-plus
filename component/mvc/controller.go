package mvc

import (
	"github.com/gin-gonic/gin"
)

// AbstractController defines the basic contract for controllers managed by the MVC framework.
// Any struct implementing this interface can be recognized and managed as a controller.
type AbstractController interface {
	// SetRoutes is used to register the controller's routes with a given Gin router group.
	// the group is a global project router group, is specific to context-path
	SetRoutes(group *gin.RouterGroup)
}

// Controller is a base struct that can be embedded into other structs to mark them as controllers.
// During application startup, any struct embedding Controller will be automatically scanned,
// registered as a bean in the IOC container, and its methods will be exposed as API endpoints.
//
// Usage:
//
//	func init() {
//	    ioc.RegisterBeanDefinition("userController", &UserController{})
//	}
//
//	type UserController struct {
//	    mvc.Controller
//	    // your fields...
//	}
//
//	func (u *UserController) SetRoutes(group *gin.RouterGroup) {
//	    group.GET("/user/:id", u.GetUser)
//	}
//
//	func (u *UserController) GetUser(ctx *gin.Context) {
//	    // Handle user retrieval
//	}
type Controller struct{}

func (c *Controller) BeanName() string {
	return ""
}

func (c *Controller) IsPrototype() bool {
	return false
}

func (c *Controller) SetRoutes(group *gin.RouterGroup) {
	// This method can be overridden by concrete controllers to register their routes.
	// The default implementation does nothing.
}
