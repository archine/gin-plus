package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/archine/gin-plus/v4/app"
	"github.com/archine/gin-plus/v4/util/reflectutil"
	"github.com/archine/gin-plus/v4/util/strutil"
	"net/http"
	"reflect"
	"time"

	"github.com/archine/gin-plus/v4/component/gplog"
	"github.com/archine/gin-plus/v4/component/mvc/router"
	"github.com/archine/gin-plus/v4/middleware"
	"github.com/gin-gonic/gin"
)

// GinServer represents a Gin-based HTTP server with additional features
type GinServer struct {
	// address is the address the server will listen on, formatted as "host:port"
	address string

	// conf holds the configuration for the Gin server
	conf *Config

	// server is the underlying HTTP server instance
	server *http.Server

	// middlewares is a slice of Gin middleware functions that will be applied globally to all routes.
	// Note: if server is running, this slice will be cleared after applying the middlewares to avoid memory leaks.
	middlewares []gin.HandlerFunc

	// routes is a slice of registered routes that will be applied to the Gin engine.
	// Note: if server is running, this slice will be cleared after applying the routes to avoid memory leaks.
	routes []*router.Route
}

// RegisterMiddleware registers a middleware to the Gin server
func (s *GinServer) RegisterMiddleware(middleware ...gin.HandlerFunc) {
	s.middlewares = append(s.middlewares, middleware...)
}

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
func (s *GinServer) RegisterRoutes(r ...*router.Route) {
	s.routes = append(s.routes, r...)
}

// GetAddress returns the address the Gin server will listen on
func (s *GinServer) GetAddress() string {
	return s.address
}

// Run starts the Gin server with the provided configuration.
func (s *GinServer) Run(appCtx *app.Context) error {
	var conf Config
	if err := appCtx.GetConfigure().Unmarshal("gin-plus.server", &conf); err != nil {
		return fmt.Errorf("failed to unmarshal gin-plus config: %w", err)
	}
	conf.Validate()

	s.address = fmt.Sprintf("%s:%d", conf.Host, conf.Port)

	gin.SetMode(conf.Mode)
	engine := gin.New()
	engine.RemoveExtraSlash = true
	engine.MaxMultipartMemory = conf.MaxMultipartMemory

	if conf.AllowedCors {
		engine.Use(middleware.Cors())
	}
	if len(s.middlewares) > 0 {
		engine.Use(s.middlewares...)
		s.middlewares = nil
	}

	err := s.applyRoute(engine, conf.ContextPath, conf.EnableHealthCheck, appCtx)
	if err != nil {
		return fmt.Errorf("failed to apply routes: %w", err)
	}

	serve := http.Server{
		Addr:                         s.address,
		Handler:                      engine,
		DisableGeneralOptionsHandler: true,
		ReadTimeout:                  conf.ReadTimeout,
		WriteTimeout:                 conf.WriteTimeout,
		ReadHeaderTimeout:            conf.ReadHeaderTimeout,
		IdleTimeout:                  conf.IdleTimeout,
	}

	errChan := make(chan error, 1)

	if conf.TLS != nil && conf.TLS.Enabled {
		if conf.TLS.CertFile == "" || conf.TLS.KeyFile == "" {
			return fmt.Errorf("TLS is enabled but cert file or key file is not provided")
		}
		serve.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12, // Ensure a minimum TLS version
		}
		go func() {
			errChan <- serve.ListenAndServeTLS(conf.TLS.CertFile, conf.TLS.KeyFile)
		}()
	} else {
		go func() {
			errChan <- serve.ListenAndServe()
		}()
	}

	select {
	case err := <-errChan:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return fmt.Errorf("failed to start server: %w", err)
		}
	case <-time.After(10 * time.Millisecond):
		// wait some time for the server to start
		s.conf = &conf
		s.server = &serve
		err = appCtx.RegisterBean("ginEngine", engine)
		if err != nil {
			return fmt.Errorf("failed to register Gin engine in IoC container: %w", err)
		}
		gplog.Debug(fmt.Sprintf("Gin engine instance is now available in the IoC container as 'ginEngine'"))
	}

	return nil
}

// Shutdown gracefully stops the Gin server
func (s *GinServer) Shutdown(closeFunc func(ctx context.Context)) error {
	if s == nil {
		return fmt.Errorf("server is not running")
	}
	if s.conf == nil {
		return fmt.Errorf("server configuration is not set")
	}

	shutdownCtx := context.Background()
	if s.conf.ShutdownTimeout > 0 {
		var cancelFunc context.CancelFunc
		shutdownCtx, cancelFunc = context.WithTimeout(shutdownCtx, s.conf.ShutdownTimeout)
		defer cancelFunc()
	}

	err := s.server.Shutdown(shutdownCtx)
	if err != nil {
		return fmt.Errorf("failed to shutdown server: %w", err)
	}

	if closeFunc != nil {
		closeCtx, cancel := context.WithTimeout(context.Background(), s.conf.ExitDelay)
		defer cancel()

		done := make(chan struct{})
		go func() {
			defer close(done)
			closeFunc(closeCtx)
		}()

		select {
		case <-done:
			// The close function completed successfully
		case <-closeCtx.Done():
			gplog.Warn(fmt.Sprintf("closeFunc timeout after %v", s.conf.ExitDelay))
		}
	}

	return nil
}

// apply attaches all APIs to the Gin engine.
// This function iterates through the registered routes and their methods,
// and registers them to the provided Gin engine under the specified gpctx path.
// It also optionally registers a health check endpoint if `enableHealth` is true.
//
// Args:
//   - engine: The Gin engine to which the routes will be applied.
//   - contextPath: The base path for the APIs, which will be prefixed to all routes.
//   - enableHealth: A boolean flag to enable or disable health check endpoints.
//   - ctx: The application context
//
// Note: this function is system-internal and should not be used directly in application code.
func (s *GinServer) applyRoute(engine *gin.Engine, contextPath string, enableHealth bool, ctx *app.Context) error {
	if len(s.routes) == 0 {
		return nil // No routes to apply
	}
	baseRouter := engine.Group(contextPath)
	ginCtxType := reflectutil.PtrOf[gin.Context]()

	if enableHealth {
		baseRouter.Any("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
	}

	for _, route := range s.routes {
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
					ctrlInstance, exist := ctx.GetBean(strutil.FirstToLower(route.Name))
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

	s.routes = nil
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
