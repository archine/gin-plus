package server

import (
	"crypto/tls"
	"fmt"
	"github.com/archine/gin-plus/v4/component/mvc/router"
	"github.com/archine/gin-plus/v4/middleware"
	"github.com/gin-gonic/gin"
	"net/http"
)

type GinServer struct {
	middlewares []gin.HandlerFunc
	address     string
}

// RegisterMiddleware registers a middleware to the Gin server
func (s *GinServer) RegisterMiddleware(middleware ...gin.HandlerFunc) {
	s.middlewares = append(s.middlewares, middleware...)
}

// GetAddress returns the address the Gin server will listen on
func (s *GinServer) GetAddress() string {
	return s.address
}

// Run starts the Gin server with the registered middlewares
func (s *GinServer) Run(conf *Config) (err error) {
	s.address = fmt.Sprintf("%s:%d", conf.Server.Host, conf.Server.Port)

	engine := gin.New()
	engine.RemoveExtraSlash = true
	engine.MaxMultipartMemory = conf.Server.MaxMultipartMemory

	if conf.Server.AllowedCors {
		engine.Use(middleware.Cors())
	}
	if len(s.middlewares) > 0 {
		engine.Use(s.middlewares...)
	}

	err = router.Apply(engine, conf.Server.ContextPath, conf.Server.EnableHealthCheck)
	if err != nil {
		return fmt.Errorf("failed to apply routes: %w", err)
	}

	serve := http.Server{
		Addr:                         s.address,
		Handler:                      engine,
		DisableGeneralOptionsHandler: true,
		ReadTimeout:                  conf.Server.ReadTimeout,
		WriteTimeout:                 conf.Server.WriteTimeout,
		ReadHeaderTimeout:            conf.Server.ReadHeaderTimeout,
		IdleTimeout:                  conf.Server.IdleTimeout,
	}
	if conf.Server.TLS != nil && conf.Server.TLS.Enabled {
		// If TLS is enabled, ensure cert and key files are provided
		if conf.Server.TLS.CertFile == "" || conf.Server.TLS.KeyFile == "" {
			return fmt.Errorf("TLS is enabled but cert file or key file is not provided")
		}
		serve.TLSConfig = &tls.Config{
			MinVersion: tls.VersionTLS12, // Ensure a minimum TLS version
		}
		go func() {
			err = serve.ListenAndServeTLS(conf.Server.TLS.CertFile, conf.Server.TLS.KeyFile)
		}()
	} else {
		err = serve.ListenAndServe()
	}

	return err
}
