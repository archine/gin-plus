package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/archine/gin-plus/v4/app"
	"github.com/archine/gin-plus/v4/component/mvc"
	"net/http"
	"reflect"
	"time"

	"github.com/archine/gin-plus/v4/component/gplog"
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
}

// RegisterMiddleware registers a middleware to the Gin server
func (s *GinServer) RegisterMiddleware(middleware ...gin.HandlerFunc) {
	s.middlewares = append(s.middlewares, middleware...)
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

	err := s.applyRoute(appCtx, engine, conf.ContextPath, conf.EnableHealthCheck)
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
// It also optionally registers a health check endpoint if `enableHealth` is true.
//
// Args:
//   - appCtx: The application context
//   - engine: The Gin engine to which the routes will be applied.
//   - contextPath: The base path for the APIs, which will be prefixed to all routes.
//   - enableHealth: A boolean flag to enable or disable health check endpoints.
//
// Note: this function is system-internal and should not be used directly in application code.
func (s *GinServer) applyRoute(appCtx *app.Context, engine *gin.Engine, contextPath string, enableHealth bool) error {
	ctrls, found := appCtx.GetAllBeansByType(reflect.TypeOf((*mvc.AbstractController)(nil)).Elem())

	if !found {
		return nil
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

	gplog.Info("All routes have been loaded successfully.")

	return nil
}
