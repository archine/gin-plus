package mvc

import (
	"github.com/archine/ioc"
	"github.com/gin-gonic/gin"
	"strings"
)

// Interface abstract top-level data structure

// Global controller cache
var controllerCache []abstractController

type abstractController interface {
	// Get current controller all api func
	getMethods() []*method

	// Sets the current controller's api method
	putMethods(apiFunc ...*method)

	// Get current controller global func
	getHandleFunc() []gin.HandlerFunc

	// PostConstruct Triggered after dependency injection is completed. You can continue to decorate the controller here
	PostConstruct()
}

// Controller Declares the structure to be a controller
// you can add api methods to it
type Controller struct {
	// methodCache
	methodCache []*method

	// controller api global handle func
	handleFunc []gin.HandlerFunc

	// Api base path
	prefix string
}

// api method struct
type method struct {
	// HttpMethod
	// Http method,such as POST, GET, DELETE, PUT...
	httpMethod string

	// Path
	// Api path, merge with the Prefix()
	path string

	// ApiFunc
	// The real function corresponding to the API
	apiFunc []gin.HandlerFunc

	// Whether to set global func
	setGlobalFunc bool
}

// ApiInfo api info,used by group method
type ApiInfo struct {
	// Path
	// Api path, merge with the Prefix()
	Path string

	// ApiFunc
	// The real function corresponding to the API
	ApiFunc gin.HandlerFunc

	// SetGlobalFunc Whether to set global func
	SetGlobalFunc bool
}

func (c *Controller) getMethods() []*method {
	return c.methodCache
}

func (c *Controller) putMethods(apiFunc ...*method) {
	c.methodCache = append(c.methodCache, apiFunc...)
}

func (c *Controller) getHandleFunc() []gin.HandlerFunc {
	return c.handleFunc
}

func (c *Controller) PostConstruct() {

}

// Prefix
// Set api base path
func (c *Controller) Prefix(path string) *Controller {
	c.prefix = path
	return c
}

// GlobalFunc
// Sets the global shared handler that applies to all apis under the current controller
// append after the func specified in the Apply method globalFunc parameter
func (c *Controller) GlobalFunc(funcs ...gin.HandlerFunc) *Controller {
	c.handleFunc = funcs
	return c
}

// Api
// Add the API to the controller
// @param httpMethod: Http method,such as POST, GET, DELETE, PUT, PATCH, OPTIONS, HEAD, support the lowercase
// @param apiPath: Api path, merge with the Prefix()
// @param apiFunc: The real function corresponding to the API
// @param setGlobalFunc: Whether to set global func
// @return this
func (c *Controller) Api(httpMethod, apiPath string, apiFunc gin.HandlerFunc, setGlobalFunc bool) *Controller {
	m := &method{
		httpMethod:    httpMethod,
		path:          c.prefix + apiPath,
		setGlobalFunc: setGlobalFunc,
		apiFunc:       []gin.HandlerFunc{apiFunc},
	}
	c.putMethods(m)
	return c
}

// ApiGroup
// Add api group to controller
// @param httMethod: Http method,such as POST, GET, DELETE, PUT, PATCH, OPTIONS, HEAD, support the lowercase
// @param apiInfos: api group
// @return this
func (c *Controller) ApiGroup(httpMethod string, apiInfos []*ApiInfo) *Controller {
	if len(apiInfos) == 0 {
		return c
	}
	methodGroup := make([]*method, len(apiInfos))
	for i, info := range apiInfos {
		m := &method{
			httpMethod:    httpMethod,
			path:          c.prefix + info.Path,
			setGlobalFunc: info.SetGlobalFunc,
			apiFunc:       []gin.HandlerFunc{info.ApiFunc},
		}
		methodGroup[i] = m
	}
	c.putMethods(methodGroup...)
	return c
}

// Post
// Add the Post API to the controller
// return this
func (c *Controller) Post(path string, apiFunc gin.HandlerFunc, setGlobalFunc bool) *Controller {
	return c.Api("POST", path, apiFunc, setGlobalFunc)
}

func (c *Controller) PostGroup(apiInfos []*ApiInfo) *Controller {
	return c.ApiGroup("POST", apiInfos)
}

// Get
// Add the Get API to the controller
// return this
func (c *Controller) Get(path string, apiFunc gin.HandlerFunc, setGlobalFunc bool) *Controller {
	return c.Api("GET", path, apiFunc, setGlobalFunc)
}

func (c *Controller) GetGroup(apiInfos []*ApiInfo) *Controller {
	return c.ApiGroup("GET", apiInfos)
}

// Delete
// Add the Delete API to the controller
// return this
func (c *Controller) Delete(path string, apiFunc gin.HandlerFunc, setGlobalFunc bool) *Controller {
	return c.Api("DELETE", path, apiFunc, setGlobalFunc)
}

func (c *Controller) DeleteGroup(apiInfos []*ApiInfo) *Controller {
	return c.ApiGroup("DELETE", apiInfos)
}

// Put
// Add the Put API to the controller
// return this
func (c *Controller) Put(path string, apiFunc gin.HandlerFunc, setGlobalFunc bool) *Controller {
	return c.Api("PUT", path, apiFunc, setGlobalFunc)
}

func (c *Controller) PutGroup(apiInfos []*ApiInfo) *Controller {
	return c.ApiGroup("PUT", apiInfos)
}

// Register controller
func Register(controller abstractController) {
	controllerCache = append(controllerCache, controller)
}

// Apply all apis to the gin engine
// @param e: gin.Engine
// @param autowired: whether enable autowired properties
// @param globalFunc: global api func, the highest priority
func Apply(e *gin.Engine, autowired bool, globalFunc ...gin.HandlerFunc) {
	for _, controller := range controllerCache {
		if autowired {
			ioc.Inject(controller)
		}
		controller.PostConstruct()
		for _, method := range controller.getMethods() {
			if method.setGlobalFunc {
				if len(controller.getHandleFunc()) > 0 {
					globalFunc = append(globalFunc, controller.getHandleFunc()...)
				}
				method.apiFunc = append(globalFunc, method.apiFunc...)
			}
			switch strings.ToUpper(method.httpMethod) {
			case "GET":
				e.GET(method.path, method.apiFunc...)
			case "POST":
				e.POST(method.path, method.apiFunc...)
			case "DELETE":
				e.DELETE(method.path, method.apiFunc...)
			case "PUT":
				e.PUT(method.path, method.apiFunc...)
			case "PATCH":
				e.PATCH(method.path, method.apiFunc...)
			case "OPTIONS":
				e.OPTIONS(method.path, method.apiFunc...)
			case "HEAD":
				e.HEAD(method.path, method.apiFunc...)
			}
		}
	}
}
