package routing

import (
	"fmt"
	"github.com/archine/gin-plus/v4/component/gplog"
	"github.com/gin-gonic/gin"
	"net/http"
)

type Method struct {
	HandlerFunc gin.HandlerFunc // Handler function for the method
	HttpMethod  string          // HTTP method (GET, POST, PUT, DELETE, etc.)
	Path        string          // URL path for the route
}

type Route struct {
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
//	        BasePath: "/user",
//	        Methods: []*Method{
//	            {Name: "GetUser", HttpMethod: "GET", Path: "/:id"},
//	            {Name: "CreateUser", HttpMethod: "POST", Path: "/"},
//	        },
//	    },
//	    &Route{
//	        BasePath: "/product",
//	        Methods: []*Method{
//	            {Name: "GetProduct", HttpMethod: "GET", Path: "/:id"},
//	            {Name: "CreateProduct", HttpMethod: "POST", Path: "/"},
//	        },
//	    },
//	)
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

	if enableHealth {
		// Register health check endpoint
		baseRouter.Any("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "healthy"})
		})
	}

	for _, route := range routes {
		var ctrlRouter *gin.RouterGroup

		if route.BasePath != "" {
			ctrlRouter = baseRouter.Group(route.BasePath)
		} else {
			ctrlRouter = baseRouter
		}
		for _, method := range route.Methods {
			if method.HandlerFunc == nil {
				gplog.Warn(fmt.Sprintf("Skipping method %s on path %s: no handler function defined", method.HttpMethod, method.Path))
				continue // Skip if no handler function is defined
			}

			// Register the method with the Gin engine
			switch method.HttpMethod {
			case http.MethodGet:
				ctrlRouter.GET(method.Path, method.HandlerFunc)
			case http.MethodPost:
				ctrlRouter.POST(method.Path, method.HandlerFunc)
			case http.MethodPut:
				ctrlRouter.PUT(method.Path, method.HandlerFunc)
			case http.MethodDelete:
				ctrlRouter.DELETE(method.Path, method.HandlerFunc)
			case http.MethodPatch:
				ctrlRouter.PATCH(method.Path, method.HandlerFunc)
			case http.MethodHead:
				ctrlRouter.HEAD(method.Path, method.HandlerFunc)
			case http.MethodOptions:
				ctrlRouter.OPTIONS(method.Path, method.HandlerFunc)
			default:
				return fmt.Errorf("unsupported HTTP method: %s", method.HttpMethod)
			}
		}
	}

	return nil
}
