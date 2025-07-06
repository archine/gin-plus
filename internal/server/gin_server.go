package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"github.com/archine/gin-plus/v4/internal/syslink"
	"net/http"
	"time"

	"github.com/archine/gin-plus/v4/component/config"
	"github.com/archine/gin-plus/v4/component/gplog"
	"github.com/archine/gin-plus/v4/component/ioc"
	"github.com/archine/gin-plus/v4/middleware"
	"github.com/gin-gonic/gin"
)

// GinServer represents a Gin-based HTTP server with additional features
type GinServer struct {
	address     string
	conf        *Config
	server      *http.Server
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
func (s *GinServer) Run(configure config.Configure) error {
	var conf Config
	if err := configure.Unmarshal("gin-plus.server", &conf); err != nil {
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

	err := syslink.ApplyRoute(engine, conf.ContextPath, conf.EnableHealthCheck)
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
		err = ioc.RegisterBean("ginEngine", engine)
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
