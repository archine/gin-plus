package mvc

import (
	"github.com/archine/ast-base/core"
	"github.com/archine/ioc"
	"github.com/gin-gonic/gin"
	"reflect"
)

// Annotations represents the annotations of an API method.
type Annotations map[string]string

// Global controller cache.
var controllerCache []abstractController

// Cache for annotations of each API.
var annotationCache map[string]Annotations

// abstractController defines the interface for a controller that
// requires a post-construction initialization method.
type abstractController interface {
	// PostConstruct is triggered after dependency injection is completed.
	// This method can be used to further initialize the controller.
	PostConstruct()
}

// Controller is a base struct that declares an entity as a controller.
// API methods can be added to this struct.
type Controller struct{}

// PostConstruct is a default implementation for Controller.
func (c *Controller) PostConstruct() {}

// Register adds controllers to the global cache.
func Register(controllers ...abstractController) {
	controllerCache = append(controllerCache, controllers...)
}

// IsController checks if a given value implements the abstractController interface.
func IsController(v interface{}) bool {
	ct := reflect.TypeOf(v)
	return ct.Kind() == reflect.Ptr && ct.Implements(reflect.TypeOf((*abstractController)(nil)).Elem())
}

// Apply attaches all APIs to the Gin engine.
// @param e: the Gin engine.
// @param autowired: if true, enables property injection via IoC.
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

		controllerType := reflect.TypeOf(controller).Elem()
		controllerValue := reflect.ValueOf(controller)
		methodInfos := core.Apis[controllerType.Name()]

		for _, m := range methodInfos {
			methodValue := controllerValue.MethodByName(m.Name)
			if methodValue.Kind() == reflect.Invalid {
				continue
			}

			ginMethod := ginProxy.MethodByName(m.Method)
			args := []reflect.Value{reflect.ValueOf(m.ApiPath), methodValue}
			ginMethod.Call(args)
			annotationCache[m.ApiPath] = m.Annotations
		}

		if len(controllerCache) == 1 {
			controllerCache = nil
			return
		}

		controllerCache = controllerCache[1:]
	}
	core.Apis = nil // Trigger garbage collection
}

// GetAnnotation retrieves the specified annotation from the current context.
// Returns the annotation value and a boolean indicating whether the annotation exists.
func GetAnnotation(ctx *gin.Context, annotationName string) (val string, has bool) {
	anno, has := annotationCache[ctx.Request.URL.Path]
	if !has || len(anno) == 0 {
		return "", false
	}
	val, has = anno[annotationName]
	return
}

// MethodInterceptor allows for pre- and post-processing of API method calls.
type MethodInterceptor interface {
	// Predicate determines whether to intercept the request.
	Predicate(ctx *gin.Context) bool

	// PreHandle is triggered before the API method is invoked.
	// If you want to abort the current request, call ctx.Abort() and handle the response inside this method.
	PreHandle(ctx *gin.Context)

	// PostHandle is triggered after the API method is invoked.
	// If you want to abort the current request, call ctx.Abort() and handle the response inside this method.
	PostHandle(ctx *gin.Context)
}
