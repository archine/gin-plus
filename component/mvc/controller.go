package mvc

import (
	"github.com/archine/gin-plus/v4/internal/event_manager"
	"github.com/archine/gin-plus/v4/ioc"
	"github.com/gin-gonic/gin"
)

// Cache for annotations of each API.
var annotationCache map[string]map[string]string

// Controller is a base struct that declares an entity as a controller.
// API methods can be added to this struct.
type Controller struct {
	ioc.Bean
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
	//var ginProxy reflect.Value
	//if len(ast_base.Result.Apis) > 0 {
	//	ginProxy = reflect.ValueOf(engine)
	//}

	//for _, controller := range controllerCache {
	//	ioc.Inject(controller)
	//	controller.PostConstruct()
	//
	//	controllerType := reflect.TypeOf(controller).Elem()
	//	controllerValue := reflect.ValueOf(controller)
	//	methodInfos := ast_base.Result.Apis[controllerType.Name()]
	//
	//	for _, m := range methodInfos {
	//		methodValue := controllerValue.MethodByName(m.Name)
	//		if methodValue.Kind() == reflect.Invalid {
	//			continue
	//		}
	//
	//		ginMethod := ginProxy.MethodByName(m.Method)
	//		if ginMethod.Kind() == reflect.Invalid {
	//			continue
	//		}
	//		args := []reflect.Value{reflect.ValueOf(m.APIPath), methodValue}
	//		ginMethod.Call(args)
	//		//annotationCache[m.APIPath] = m.Annotations
	//	}
	//}
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
