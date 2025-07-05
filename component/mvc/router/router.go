package router

import (
	"fmt"
	"github.com/archine/gin-plus/v4/component/ioc"
	"github.com/archine/gin-plus/v4/util/strutil"
	"net/http"
	"reflect"

	"github.com/archine/gin-plus/v4/component/gplog"
	"github.com/archine/gin-plus/v4/util/reflectutil"
	"github.com/gin-gonic/gin"
)

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
//	        Name: "userController",
//	        BasePath: "/api/v1",
//	        Methods: []*Method{
//	            {HttpMethod: "GET", Path: "/users", NameOrFunc: getUsers},
//	            {HttpMethod: "POST", Path: "/users", NameOrFunc: "CreateUser"},
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

// apply attaches all APIs to the Gin engine.
// This function iterates through the registered routes and their methods,
// and registers them to the provided Gin engine under the specified context path.
// It also optionally registers a health check endpoint if `enableHealth` is true.
//
// Args:
//   - engine: The Gin engine to which the routes will be applied.
//   - contextPath: The base path for the APIs, which will be prefixed to all routes.
//   - enableHealth: A boolean flag to enable or disable health check endpoints.
//
// Note: this function is system-internal and should not be used directly in application code.
func apply(engine *gin.Engine, contextPath string, enableHealth bool) error {
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
		var ctrlValue reflect.Value
		var ctrlRouter *gin.RouterGroup

		if route.BasePath != "" {
			ctrlRouter = baseRouter.Group(route.BasePath)
		} else {
			ctrlRouter = baseRouter
		}

		for _, method := range route.Methods {
			var handler gin.HandlerFunc
			var err error

			switch v := method.NameOrFunc.(type) {
			case string:
				// If the method name is a string, we assume it's the name of a method in the controller.
				if v == "" {
					gplog.Warn(fmt.Sprintf("Method name is empty for route %s, path=%s, skipping registration", route.Name, method.Path))
					continue
				}
				if !ctrlValue.IsValid() {
					ctrlInstance, exist := ioc.GetBean(strutil.FirstToLower(route.Name))
					if !exist {
						gplog.Warn(fmt.Sprintf("Controller %s not found in IOC container, skipping route registration", route.Name))
						continue
					}
					ctrlValue = reflect.ValueOf(ctrlInstance)
				}

				handler, err = convertMethodToHandler(ctrlValue, v, route.Name, ginCtxType)
				if err != nil {
					gplog.Warn(fmt.Sprintf("Failed to convert method %s.%s: %v", route.Name, v, err))
					continue
				}

			case gin.HandlerFunc:
				handler = v
			default:
				gplog.Warn(fmt.Sprintf("Invalid NameOrFunc type for route %s, path=%s, skipping registration", route.Name, method.Path))
				continue
			}

			ctrlRouter.Handle(method.HttpMethod, method.Path, handler)
		}
	}

	routes = nil
	gplog.Info("All routes applied successfully...")

	return nil
}

// convertMethodToHandler converts a method of a controller to a gin.HandlerFunc.
func convertMethodToHandler(ctrlValue reflect.Value, methodName, ctrlName string, ginCtxType reflect.Type) (gin.HandlerFunc, error) {
	methodValue := ctrlValue.MethodByName(methodName)
	if !methodValue.IsValid() {
		return nil, fmt.Errorf("method %s not found in controller %s", methodName, ctrlName)
	}

	if methodValue.Type().NumIn() != 1 || methodValue.Type().In(0) != ginCtxType {
		return nil, fmt.Errorf("method %s.%s must accept exactly one *gin.Context parameter", ctrlName, methodName)
	}

	handlerFunc, ok := methodValue.Interface().(func(*gin.Context))
	if !ok {
		return nil, fmt.Errorf("method %s.%s cannot be converted to gin.HandlerFunc", ctrlName, methodName)
	}

	return handlerFunc, nil

}
