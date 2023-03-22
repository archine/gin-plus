package mvc

import (
	"github.com/archine/gin-plus/v2/ast"
	"github.com/archine/ioc"
	"github.com/gin-gonic/gin"
	"net/http"
	"reflect"
)

// Interface abstract top-level data structure

// Global controller cache
var controllerCache []abstractController

type abstractController interface {
	// PostConstruct Triggered after dependency injection is completed. You can continue to decorate the controller here
	PostConstruct()
}

// Controller Declares the structure to be a controller
// you can add api methods to it
type Controller struct{}

func (c *Controller) PostConstruct() {}

// Register controllers
func Register(controller ...abstractController) {
	controllerCache = append(controllerCache, controller...)
}

// IsController Determine whether it is controller
func IsController(v interface{}) bool {
	ct := reflect.TypeOf(v)
	if ct.Kind() != reflect.Ptr {
		return false
	}
	return ct.Implements(reflect.TypeOf((*abstractController)(nil)).Elem())
}

// Apply all apis to the gin engine
// @param e: gin.Engine
// @param autowired: whether enable autowired properties
func Apply(e *gin.Engine, autowired bool) {
	if ast.Apis == nil {
		for _, controller := range controllerCache {
			if autowired {
				ioc.Inject(controller)
			}
		}
		return
	}
	ginProxy := reflect.ValueOf(e)
	for _, controller := range controllerCache {
		if autowired {
			ioc.Inject(controller)
		}
		controller.PostConstruct()
		controllerTypeOf := reflect.TypeOf(controller)
		controllerProxy := reflect.ValueOf(controller)
		for i := 0; i < controllerTypeOf.NumMethod(); i++ {
			methodProxy := controllerTypeOf.Method(i)
			methodFullName := controllerTypeOf.Elem().Name() + "/" + methodProxy.Name
			if info, ok := ast.Apis[methodFullName]; ok {
				for _, methodInfo := range info {
					ginMethod := ginProxy.MethodByName(methodInfo.Method)
					args := []reflect.Value{reflect.ValueOf(methodInfo.ApiPath)}
					args = append(args, controllerProxy.MethodByName(methodProxy.Name))
					ginMethod.Call(args)
				}
				delete(ast.Apis, methodFullName)
			}
		}
		if len(controllerCache) == 1 {
			controllerCache = nil
			return
		}
		controllerCache = controllerCache[1:]
	}
}

// MethodInterceptor API method interceptor.
// You can do logical processing before and after method calls
type MethodInterceptor interface {
	// Predicate true means intercept
	Predicate(request *http.Request) bool

	// PreHandle triggered before method invocation.
	// if you want to abort the current request, just call abort() and response inside the method
	PreHandle(ctx *gin.Context)

	// PostHandle triggered after method invocation.
	// if you want to abort the current request, just call abort() and response inside the method
	PostHandle(ctx *gin.Context)
}
