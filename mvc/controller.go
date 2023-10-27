package mvc

import (
	"github.com/archine/ast-base/core"
	"github.com/archine/ioc"
	"github.com/gin-gonic/gin"
	"reflect"
)

// Annotations the annotation of Api method
type Annotations map[string]string

// Global controller cache
var controllerCache []abstractController

// Annotations of each API
var annotationCache map[string]Annotations

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
	if core.Apis == nil {
		for _, controller := range controllerCache {
			if autowired {
				ioc.Inject(controller)
			}
		}
		return
	}
	ginProxy := reflect.ValueOf(e)
	annotationCache = make(map[string]Annotations)
	for _, controller := range controllerCache {
		if autowired {
			ioc.Inject(controller)
		}
		controller.PostConstruct()
		controllerTypeOf := reflect.TypeOf(controller).Elem()
		controllerProxy := reflect.ValueOf(controller)
		methodInfosAst := core.Apis[controllerTypeOf.Name()]
		for _, m := range methodInfosAst {
			mValueProxy := controllerProxy.MethodByName(m.Name)
			if mValueProxy.Kind() == reflect.Invalid {
				continue
			}
			ginMethod := ginProxy.MethodByName(m.Method)
			args := []reflect.Value{reflect.ValueOf(m.ApiPath)}
			args = append(args, mValueProxy)
			ginMethod.Call(args)
			annotationCache[m.ApiPath] = m.Annotations
		}
		if len(controllerCache) == 1 {
			controllerCache = nil
			return
		}
		controllerCache = controllerCache[1:]
	}
	core.Apis = nil // GC
}

// GetAnnotation Gets the specified annotation
// Returns the value of this annotation, when the has is false mine this val is empty
func GetAnnotation(ctx *gin.Context, annotationName string) (val string, has bool) {
	anno, has := annotationCache[ctx.FullPath()]
	if !has || len(anno) == 0 {
		return "", false
	}
	val, has = anno[annotationName]
	return
}

// MethodInterceptor API method interceptor
// You can do logical processing before and after method calls
type MethodInterceptor interface {
	// Predicate true means intercept
	Predicate(ctx *gin.Context) bool

	// PreHandle triggered before method invocation
	// if you want to abort the current request, just call abort() and response inside the method
	PreHandle(ctx *gin.Context)

	// PostHandle triggered after method invocation
	// if you want to abort the current request, just call abort() and response inside the method
	PostHandle(ctx *gin.Context)
}
