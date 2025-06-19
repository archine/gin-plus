package mvc

import (
	"github.com/archine/ast-base"
	"github.com/archine/gin-plus/v3/internal/event_manager"
	"github.com/archine/gin-plus/v3/ioc"
	"github.com/gin-gonic/gin"
	"reflect"
)

// Global controller cache.
var controllerCache []abstractController

// Cache for annotations of each API.
var annotationCache map[string]map[string]string

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

// SetAnnotations sets the annotations for a specific API path.
func SetAnnotations(annos map[string]map[string]string) {
	annotationCache = annos
}

// IsController checks if a given value implements the abstractController interface.
func IsController(v interface{}) bool {
	ct := reflect.TypeOf(v)
	return ct.Kind() == reflect.Ptr && ct.Implements(reflect.TypeOf((*abstractController)(nil)).Elem())
}

// Apply attaches all APIs to the Gin engine.
// It uses reflection to dynamically bind methods to the Gin engine based on the API definitions
// provided by the ast_base.Result.Apis map.
// It also injects dependencies into the controllers
//
// Args:
//
//	engine: The Gin engine to which the APIs will be attached.
//	eventManager: The event manager to handle application events.
func Apply(engine *gin.Engine, eventManager *event_manager.AppEventManager) {
	var ginProxy reflect.Value
	if len(ast_base.Result.Apis) > 0 {
		ginProxy = reflect.ValueOf(engine)
	}

	for _, controller := range controllerCache {
		ioc.Inject(controller)
		controller.PostConstruct()

		controllerType := reflect.TypeOf(controller).Elem()
		controllerValue := reflect.ValueOf(controller)
		methodInfos := ast_base.Result.Apis[controllerType.Name()]

		for _, m := range methodInfos {
			methodValue := controllerValue.MethodByName(m.Name)
			if methodValue.Kind() == reflect.Invalid {
				continue
			}

			ginMethod := ginProxy.MethodByName(m.Method)
			if ginMethod.Kind() == reflect.Invalid {
				continue
			}
			args := []reflect.Value{reflect.ValueOf(m.APIPath), methodValue}
			ginMethod.Call(args)
			//annotationCache[m.APIPath] = m.Annotations
		}
	}

	ast_base.Result = nil
}

// GetAnnotation retrieves the specified annotation from the current context.
// Returns the annotation value and a boolean indicating whether the annotation exists.
func GetAnnotation(ctx *gin.Context, annotationName string) (val string, has bool) {
	anno, has := annotationCache[ctx.FullPath()]
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
