package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"net/http"
	"time"

	"github.com/archine/gin-plus/v4/component/gplog"
	"github.com/archine/gin-plus/v4/component/ioc"
	"github.com/archine/gin-plus/v4/component/mvc"
	"github.com/archine/gin-plus/v4/internal/vars/sysconf"
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
func (s *GinServer) Init() error {
	var conf Config
	if err := sysconf.Provider.Unmarshal("gin-plus.server", &conf); err != nil {
		return err
	}

	conf.Validate()
	s.conf = &conf

	return nil
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
func (s *GinServer) Run() error {
	gin.SetMode(s.conf.Mode)
	engine := gin.New()
	engine.RemoveExtraSlash = true
	engine.MaxMultipartMemory = s.conf.MaxMultipartMemory

	if !s.conf.DisableDefaultLogger {
		logConf := gin.LoggerConfig{
			Output:    &logWriter{},
			SkipPaths: s.conf.SkipLogPaths,
		}
		if gplog.GetLogger().GetFormat() == gplog.JSONFormat {
			logConf.Formatter = jsonFormatter
		} else {
			logConf.Formatter = consoleFormatter
		}

		engine.Use(gin.LoggerWithConfig(logConf))
	}

	if !s.conf.DisableDefaultCors {
		engine.Use(middleware.CORS())
	}

	if !s.conf.DisableDefaultRecovery {
		engine.Use(middleware.Recovery())
	}

	if len(s.middlewares) > 0 {
		engine.Use(s.middlewares...)
		s.middlewares = nil
	}

	s.applyRoute(engine, s.conf.ContextPath, s.conf.EnableHealthCheck)

	serve := http.Server{
		Addr:                         fmt.Sprintf("%s:%d", s.conf.Host, s.conf.Port),
		Handler:                      engine,
		DisableGeneralOptionsHandler: s.conf.DisablePassOptions,
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
	case <-time.After(3 * time.Millisecond):
		// wait a moment to ensure the server has started before logging
		s.server = &serve
	}

	return nil
}

// Shutdown gracefully stops the Gin server
func (s *GinServer) Shutdown(closeFunc func(ctx context.Context)) error {
	if s.server == nil {
		return fmt.Errorf("server is not running")
	}
	if s.conf == nil {
		return fmt.Errorf("server configuration is not initialized")
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

	s.server = nil

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
func (s *GinServer) applyRoute(engine *gin.Engine, contextPath string, enableHealth bool) {
	baseRouter := engine.Group(contextPath)

	if enableHealth {
		baseRouter.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
	}

	ctrls := ioc.GetAllBeansByType[mvc.AbstractController]()

	if ctrls == nil {
		return
	}

	for _, ctrl := range ctrls {
		ctrl.SetRoutes(baseRouter)
	}

	gplog.Info(fmt.Sprintf("All routes have been applied to the Gin engine, parsed %d controllers", len(ctrls)))
	gplog.Info(fmt.Sprintf("Application run with context path: '%s'", contextPath))
}
