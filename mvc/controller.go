package mvc

import (
	"github.com/archine/gin-plus/v2/ast"
	"github.com/archine/ioc"
	"github.com/gin-gonic/gin"
	"reflect"
)

// Interface abstract top-level data structure

// Global controller cache
var controllerCache []abstractController

type abstractController interface {
	// PostConstruct Triggered after dependency injection is completed. You can continue to decorate the controller here
	PostConstruct()

	// CallBefore local function, scoped to the current controller, fired before calling the API
	// @funcName API func name
	CallBefore(funcName string) []gin.HandlerFunc
}

// Controller Declares the structure to be a controller
// you can add api methods to it
type Controller struct {
}

func (c *Controller) PostConstruct() {

}

func (c *Controller) CallBefore(funcName string) []gin.HandlerFunc {
	return nil
}

// Register controller
func Register(controller abstractController) {
	controllerCache = append(controllerCache, controller)
}

// Apply all apis to the gin engine
// @param e: gin.Engine
// @param autowired: whether enable autowired properties
// @param globalFunc: global api func, the highest priority
// @param controllerDir: controller file directory
func Apply(e *gin.Engine, autowired bool, apiInfo map[string][]*ast.MethodInfo, globalFunc ...gin.HandlerFunc) {
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
			if info, ok := apiInfo[methodProxy.Name]; ok {
				controllerFuncs := controller.CallBefore(methodProxy.Name)
				for _, methodInfo := range info {
					ginMethod := ginProxy.MethodByName(methodInfo.Method)
					if !ginMethod.IsValid() {
						panic("invalid gin method: " + methodInfo.Method)
					}
					args := []reflect.Value{reflect.ValueOf(methodInfo.ApiPath)}
					if methodInfo.GlobalFunc {
						for _, handlerFunc := range globalFunc {
							args = append(args, reflect.ValueOf(handlerFunc))
						}
					}
					if controllerFuncs != nil {
						for _, handlerFunc := range controllerFuncs {
							args = append(args, reflect.ValueOf(handlerFunc))
						}
					}
					args = append(args, controllerProxy.MethodByName(methodProxy.Name))
					ginMethod.Call(args)
				}
				delete(apiInfo, methodProxy.Name)
			}
		}
	}
	controllerCache = nil
}
