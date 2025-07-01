package router

import (
	"fmt"
	"net/http"
	"reflect"

	"github.com/archine/gin-plus/v4/component/gplog"
	"github.com/archine/gin-plus/v4/component/ioc"
	"github.com/archine/gin-plus/v4/util/reflectutil"
	"github.com/archine/gin-plus/v4/util/strutil"
	"github.com/gin-gonic/gin"
)

type Method struct {
	Name       string // Name of the method (controller action)
	HttpMethod string // HTTP method (GET, POST, PUT, DELETE, etc.)
	Path       string // URL path for the route
}

type Route struct {
	Name     string    // Name of the controller
	BasePath string    // Base path for the controller
	Methods  []*Method // List of methods associated with the controller
}

var routes []*Route

// RegisterRoutes registers multiple routes to the global routes slice.
// This function allows you to add multiple routes at once, which can be useful for bulk registration
// of API endpoints.
//
// Args:
//   - routes: A variadic parameter that accepts multiple Route pointers.
//     This allows you to pass any number of Route instances to be registered.
//
// Example usage:
//
//	RegisterRoutes(
//	    &Route{
//	        BasePath: "/api/v1",
//	        Methods: []*Method{
//	            {HttpMethod: "GET", Path: "/users", HandlerFunc: getUsers},
//	            {HttpMethod: "POST", Path: "/users", HandlerFunc: createUser},
//	        },
//	    }
//	)
//
// Note: This function does not perform any validation on the routes or methods.
// It is assumed that the provided routes are valid and correctly defined.
// If a method does not have a handler function defined,
// it will be skipped with a warning logged.
// If you need to ensure that all methods have handlers, consider adding validation logic before calling this function.
func RegisterRoutes(r ...*Route) {
	routes = append(routes, r...)
}

// Apply attaches all APIs to the Gin engine.
// This function iterates through the registered routes and their methods,
// and registers them to the provided Gin engine under the specified context path.
// It also optionally registers a health check endpoint if `enableHealth` is true.
//
// Args:
//   - engine: The Gin engine to which the routes will be applied.
//   - contextPath: The base path for the APIs, which will be prefixed to all routes.
//   - enableHealth: A boolean flag to enable or disable health check endpoints.
func Apply(engine *gin.Engine, contextPath string, enableHealth bool) error {
	if len(routes) == 0 {
		return nil // No routes to apply
	}
	baseRouter := engine.Group(contextPath)
	ginCtxType := reflectutil.PtrOf[gin.Context]()

	if enableHealth {
		// Register health check endpoint
		baseRouter.Any("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
	}

	for _, route := range routes {
		ctrlInstance, exist := ioc.GetBean(strutil.FirstToLower(route.Name))
		if !exist {
			gplog.Warn(fmt.Sprintf("Controller %s not found in IOC container, skipping route registration", route.Name))
			continue
		}
		ctrlValue := reflect.ValueOf(ctrlInstance)

		var ctrlRouter *gin.RouterGroup
		if route.BasePath != "" {
			ctrlRouter = baseRouter.Group(route.BasePath)
		} else {
			ctrlRouter = baseRouter
		}

		for _, method := range route.Methods {
			if method.Name == "" {
				gplog.Warn(fmt.Sprintf("Method name is empty for route %s ,path=%s, skipping registration", route.Name, method.Path))
				continue
			}

			methodValue := ctrlValue.MethodByName(method.Name)
			if !methodValue.IsValid() {
				gplog.Warn(fmt.Sprintf("Method %s not found in controller %s, skipping registration", method.Name, route.Name))
				continue
			}

			if methodValue.Type().NumIn() != 1 || methodValue.Type().In(0) != ginCtxType {
				gplog.Warn(fmt.Sprintf("Method %s in controller %s must accept exactly one *gin.Context parameter", method.Name, route.Name))
				continue
			}

			if handleFunc, ok := methodValue.Interface().(func(*gin.Context)); ok {
				ctrlRouter.Handle(method.HttpMethod, method.Path, handleFunc)
			} else {
				gplog.Warn(fmt.Sprintf("Method %s in controller %s does not have a valid handler function signature", method.Name, route.Name))
			}
		}
	}

	routes = nil // Clear routes after applying to avoid duplicate registrations
	gplog.Info("All routes applied successfully...")

	return nil
}
