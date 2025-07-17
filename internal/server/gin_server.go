package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/archine/gin-plus/v4/app"
	"github.com/archine/gin-plus/v4/component/mvc"
	"github.com/archine/gin-plus/v4/internal/vars/sysconf"
	"net/http"
	"os"
	"reflect"
	"time"

	"github.com/archine/gin-plus/v4/component/log"
	"github.com/archine/gin-plus/v4/middleware"
	"github.com/gin-gonic/gin"
)

// GinServer represents a Gin-based HTTP server with additional features
type GinServer struct {
	// conf holds the configuration for the Gin server
	conf *Config

	// server is the underlying HTTP server instance
	server *http.Server

	// middlewares is a slice of Gin middleware functions that will be applied globally to all routes.
	// Note: if server is running, this slice will be cleared after applying the middlewares to avoid memory leaks.
	middlewares []gin.HandlerFunc
}

// NewGinServer creates a new instance of GinServer with the provided configuration.
// It initializes the server with the specified address and configuration.
func NewGinServer() *GinServer {
	return &GinServer{}
}

// Init initializes the Gin server
func (s *GinServer) Init() {
	var conf Config
	if err := sysconf.GlobalProvider.Unmarshal("gin-plus.server", &conf); err != nil {
		log.Fatal(fmt.Sprintf("Starting Gin-Engine failure with PID %d", os.Getpid()))
	}
	conf.Validate()
	s.conf = &conf
}

// RegisterMiddleware registers a middleware to the Gin server
func (s *GinServer) RegisterMiddleware(middleware ...gin.HandlerFunc) {
	s.middlewares = append(s.middlewares, middleware...)
}

// GetAddress returns the address the Gin server will listen on
func (s *GinServer) GetAddress() string {
	return fmt.Sprintf("%s:%d", s.conf.Host, s.conf.Port)
}

// GetName returns the name of the Gin server
func (s *GinServer) GetName() string {
	return s.conf.Name
}

// Run starts the Gin server with the provided configuration.
func (s *GinServer) Run(appCtx app.ApplicationContext) error {
	gin.SetMode(s.conf.Mode)
	engine := gin.New()
	engine.RemoveExtraSlash = true
	engine.MaxMultipartMemory = s.conf.MaxMultipartMemory

	if s.conf.AllowedCors {
		engine.Use(middleware.Cors())
	}
	if len(s.middlewares) > 0 {
		engine.Use(s.middlewares...)
		s.middlewares = nil
	}

	s.applyRoute(appCtx, engine, s.conf.ContextPath, s.conf.EnableHealthCheck)

	serve := http.Server{
		Addr:                         fmt.Sprintf("%s:%d", s.conf.Host, s.conf.Port),
		Handler:                      engine,
		DisableGeneralOptionsHandler: true,
		ReadTimeout:                  s.conf.ReadTimeout,
		WriteTimeout:                 s.conf.WriteTimeout,
		ReadHeaderTimeout:            s.conf.ReadHeaderTimeout,
		IdleTimeout:                  s.conf.IdleTimeout,
	}

	errChan := make(chan error, 1)

	if s.conf.TLS != nil && s.conf.TLS.Enable {
		if s.conf.TLS.CertFile == "" || s.conf.TLS.KeyFile == "" {
			return fmt.Errorf("TLS is enabled but cert file or key file is not provided")
		}
		serve.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12, // Ensure a minimum TLS version
		}
		go func() {
			errChan <- serve.ListenAndServeTLS(s.conf.TLS.CertFile, s.conf.TLS.KeyFile)
		}()
	} else {
		go func() {
			errChan <- serve.ListenAndServe()
		}()
	}

	select {
	case err := <-errChan:
		if err != nil && !errors.Is(err, http.ErrServerClosed) {
			return err
		}
	case <-time.After(10 * time.Millisecond):
		// wait some time for the server to start
		s.server = &serve
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
		return err
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
			log.Warn(fmt.Sprintf("closeFunc timeout after %v", s.conf.ExitDelay))
		}
	}

	return nil
}

// apply attaches all APIs to the Gin engine.
// This function iterates through the registered routes and their methods,
// It also optionally registers a health check endpoint if `enableHealth` is true.
//
// Args:
//   - appCtx: The application context
//   - engine: The Gin engine to which the routes will be applied.
//   - contextPath: The base path for the APIs, which will be prefixed to all routes.
//   - enableHealth: A boolean flag to enable or disable health check endpoints.
//
// Note: this function is system-internal and should not be used directly in application code.
func (s *GinServer) applyRoute(appCtx app.ApplicationContext, engine *gin.Engine, contextPath string, enableHealth bool) {
	ctrls, found := appCtx.GetAllBeansByType(reflect.TypeOf((*mvc.AbstractController)(nil)).Elem())

	if !found {
		return
	}

	baseRouter := engine.Group(contextPath)

	if enableHealth {
		baseRouter.Any("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
	}

	for _, ctrl := range ctrls {
		ctrl.(mvc.AbstractController).SetRoutes(baseRouter)
	}

	log.Info("API route registration completed: all routes are mapped and active")
}
