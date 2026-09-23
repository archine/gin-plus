package server

import (
	"context"
	"crypto/tls"
	"errors"
	"fmt"
	"log"
	"net"
	"net/http"
	"strconv"
	"time"

	"github.com/archine/gin-plus/v4/component/config"
	"github.com/archine/gin-plus/v4/component/gplog"
	"github.com/archine/gin-plus/v4/component/ioc"
	"github.com/archine/gin-plus/v4/component/mvc"
	"github.com/archine/gin-plus/v4/middleware"
	"github.com/gin-gonic/gin"
)

// GinServer represents a Gin-based HTTP server with additional features.
// Lifecycle: Init → Start → Shutdown, each invoked once by the application.
type GinServer struct {
	// conf holds the configuration for the Gin server
	conf *Config

	// server is the underlying HTTP server instance, nil until Start succeeds
	server *http.Server

	// middlewares is a slice of Gin middleware functions that will be applied globally to all routes.
	// It is released once applied to the engine.
	middlewares []gin.HandlerFunc
}

// NewGinServer creates an uninitialized GinServer; call Init before Start.
func NewGinServer() *GinServer {
	return &GinServer{}
}

// Init loads and validates the server configuration from "gin-plus.server".
func (s *GinServer) Init(cp config.Provider) error {
	var conf Config
	if err := cp.Unmarshal("gin-plus.server", &conf); err != nil {
		return fmt.Errorf("load server config: %w", err)
	}
	if err := conf.normalize(); err != nil {
		return fmt.Errorf("invalid server config: %w", err)
	}

	s.conf = &conf
	return nil
}

// RegisterMiddleware registers global middlewares; they run after the built-in ones, in order.
func (s *GinServer) RegisterMiddleware(middleware ...gin.HandlerFunc) {
	s.middlewares = append(s.middlewares, middleware...)
}

// Address returns the address the server listens on (IPv6 hosts are bracketed).
func (s *GinServer) Address() string {
	return net.JoinHostPort(s.conf.Host, strconv.Itoa(s.conf.Port))
}

// Name returns the configured application name.
func (s *GinServer) Name() string {
	return s.conf.Name
}

// ExitDelay returns the time budget for post-shutdown cleanup callbacks.
func (s *GinServer) ExitDelay() time.Duration {
	return s.conf.ExitDelay
}

// Start binds the listening socket synchronously and serves requests in the background.
//
// Binding errors (port in use, bad TLS key pair, ...) are returned directly, so a nil error
// means the server is really accepting connections. Errors raised later by the serving loop
// are delivered through the returned channel; it never receives anything on a regular Shutdown.
func (s *GinServer) Start() (<-chan error, error) {
	srv := &http.Server{
		Handler:           s.newEngine(),
		ReadTimeout:       s.conf.ReadTimeout,
		WriteTimeout:      s.conf.WriteTimeout,
		ReadHeaderTimeout: s.conf.ReadHeaderTimeout,
		IdleTimeout:       s.conf.IdleTimeout,
		// Route net/http internal errors (TLS handshake failures, handler panics, ...) into gplog.
		ErrorLog: log.New(logWriter(gplog.Warn), "", 0),
	}

	useTLS := s.conf.TLS != nil && s.conf.TLS.Enable
	if useTLS {
		// Load the key pair up front so that a bad certificate fails start-up synchronously.
		cert, err := tls.LoadX509KeyPair(s.conf.TLS.CertFile, s.conf.TLS.KeyFile)
		if err != nil {
			return nil, fmt.Errorf("load TLS key pair: %w", err)
		}
		srv.TLSConfig = &tls.Config{MinVersion: tls.VersionTLS12, Certificates: []tls.Certificate{cert}}
	}

	ln, err := net.Listen("tcp", s.Address())
	if err != nil {
		return nil, err
	}
	s.server = srv

	errCh := make(chan error, 1)
	go func() {
		var err error
		if useTLS {
			// Empty file names: certificates come from TLSConfig. ServeTLS also enables HTTP/2.
			err = srv.ServeTLS(ln, "", "")
		} else {
			err = srv.Serve(ln)
		}
		// Serve always returns a non-nil error; ErrServerClosed only reports a regular Shutdown.
		if !errors.Is(err, http.ErrServerClosed) {
			errCh <- err
		}
	}()

	return errCh, nil
}

// Shutdown gracefully stops the server within ShutdownTimeout. When the timeout elapses,
// remaining connections are closed forcibly and the timeout error is returned.
// It is a no-op if the server has never been started.
func (s *GinServer) Shutdown() error {
	if s.server == nil {
		return nil
	}

	ctx := context.Background()
	if s.conf.ShutdownTimeout > 0 {
		var cancel context.CancelFunc
		ctx, cancel = context.WithTimeout(ctx, s.conf.ShutdownTimeout)
		defer cancel()
	}

	if err := s.server.Shutdown(ctx); err != nil {
		_ = s.server.Close()
		return fmt.Errorf("graceful shutdown: %w", err)
	}
	return nil
}

// newEngine builds the Gin engine: built-in middlewares (logger → recovery),
// then user middlewares, then routes.
func (s *GinServer) newEngine() *gin.Engine {
	gin.SetMode(s.conf.Mode)
	engine := gin.New()
	engine.RemoveExtraSlash = true
	// Let *gin.Context fall back to Request.Context() for Value/Deadline/Done/Err, so passing it
	// as a context.Context sees values set upstream (e.g. trace ids) and client cancellation.
	engine.ContextWithFallback = true
	engine.MaxMultipartMemory = s.conf.MaxMultipartMemory

	if !s.conf.DisableDefaultLogger {
		logConf := gin.LoggerConfig{
			Output:    logWriter(gplog.Info),
			SkipPaths: s.conf.SkipLogPaths,
			Formatter: consoleFormatter,
		}
		if gplog.GetLogger().GetFormat() == gplog.JSONFormat {
			logConf.Formatter = jsonFormatter
		}
		engine.Use(gin.LoggerWithConfig(logConf))
	}

	if !s.conf.DisableDefaultRecovery {
		engine.Use(middleware.Recovery())
	}

	if len(s.middlewares) > 0 {
		engine.Use(s.middlewares...)
		s.middlewares = nil
	}

	s.applyRoutes(engine)
	return engine
}

// applyRoutes registers the optional health check and the routes of every controller bean,
// in bean registration order, under the configured context path.
func (s *GinServer) applyRoutes(engine *gin.Engine) {
	baseRouter := engine.Group(s.conf.ContextPath)

	if s.conf.EnableHealthCheck {
		baseRouter.GET("/health", func(c *gin.Context) {
			c.JSON(http.StatusOK, gin.H{"status": "ok"})
		})
	}

	ctrls := ioc.GetAllBeansByType[mvc.AbstractController]()
	for _, ctrl := range ctrls {
		ctrl.SetRoutes(baseRouter)
	}

	gplog.Info(fmt.Sprintf("Applied routes of %d controllers under context path '%s'", len(ctrls), s.conf.ContextPath))
}
