package mvc

import (
	"github.com/gin-gonic/gin"
)

// AbstractController defines the basic contract for controllers managed by the MVC framework.
// Any struct implementing this interface can be recognized and managed as a controller.
type AbstractController interface {
	// SetRoutes is used to register the controller's routes with a given Gin router group.
	// the group is a root group, is specific to context-path
	SetRoutes(rootGroup *gin.RouterGroup)
}
