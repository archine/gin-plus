package server

import (
	"time"
)

// Config contains HTTP server related configuration options.
type Config struct {
	// Name is the application name.
	Name string `mapstructure:"name"`

	// Port specifies the HTTP server listening port.
	// Default: 4006
	Port int `mapstructure:"port"`

	// Host specifies the HTTP server binding address.
	// Use "0.0.0.0" to bind to all interfaces, "127.0.0.1" for localhost only.
	// Default: "0.0.0.0"
	Host string `mapstructure:"host"`

	// ContextPath is the base path for the HTTP server.
	// All routes will be prefixed with this path.
	// If empty, the server will listen on the root path ("/").
	ContextPath string `mapstructure:"context-path"`

	// Mode specifies the Gin mode: "debug", "release", or "test".
	// In debug mode, Gin provides more detailed logging and error information.
	// Default: "debug"
	Mode string `mapstructure:"mode"`

	// AllowedCors enables Cross-Origin Resource Sharing (CORS) support.
	// When enabled, the server automatically adds default CORS middleware to handle cross-origin requests.
	// Default: false
	AllowedCors bool `mapstructure:"allowed-cors"`

	// MaxMultipartMemory sets the maximum memory (in bytes) for multipart form parsing.
	// Files larger than this limit are written to temporary files instead of being stored in memory.
	// This prevents excessive memory consumption when handling large file uploads.
	// Default: 8MB (8388608 bytes)
	MaxMultipartMemory int64 `mapstructure:"max-multipart-memory"`

	// WriteTimeout is the maximum duration before timing out writes of the response.
	// The timer is reset whenever a new request's header is read.
	// Unlike per-request timeouts, this applies globally to all handlers.
	// A zero or negative value disables the timeout.
	// Default: 0 (no timeout)
	WriteTimeout time.Duration `mapstructure:"write-timeout"`

	// ReadTimeout is the maximum duration for reading the entire request, including the body.
	// This timeout applies to the complete request reading process.
	// A zero or negative value disables the timeout.
	// Note: Most applications should prefer ReadHeaderTimeout for better control.
	// Default: 0 (no timeout)
	ReadTimeout time.Duration `mapstructure:"read-timeout"`

	// ReadHeaderTimeout is the maximum duration allowed to read request headers.
	// After reading headers, the connection's read deadline is reset and handlers
	// can make per-request decisions about body reading timeouts.
	// If zero, the ReadTimeout value is used. If both are zero, there is no timeout.
	// Default: 0 (uses ReadTimeout)
	ReadHeaderTimeout time.Duration `mapstructure:"read-header-timeout"`

	// IdleTimeout is the maximum duration to wait for the next request when keep-alives are enabled.
	// This helps free up resources from idle connections.
	// If zero, the ReadTimeout value is used. If both are zero, there is no timeout.
	// Default: 0 (uses ReadTimeout)
	IdleTimeout time.Duration `mapstructure:"idle-timeout"`

	// ShutdownTimeout specifies the maximum duration to wait for graceful server shutdown.
	// During this period, the server will attempt to finish processing ongoing requests
	// before forcefully terminating. A zero value means no timeout (wait indefinitely).
	// Default: 0 (no timeout)
	ShutdownTimeout time.Duration `mapstructure:"shutdown-timeout"`

	// ExitDelay is the duration to wait before the application process exits after shutdown.
	// This grace period allows for final cleanup tasks, log flushing, or external notifications.
	// Useful for ensuring all resources are properly released before process termination.
	// Default: 3s
	ExitDelay time.Duration `mapstructure:"exit-delay"`

	// EnableHealthCheck enables the built-in health check endpoint.
	// When enabled, the server automatically registers a health check route at /health.
	// This endpoint can be used to monitor the application's health status.
	// Default: false
	EnableHealthCheck bool `mapstructure:"enable-health-check"`

	// TLS configuration for HTTPS support
	TLS *TLSConfig `mapstructure:"tls"`
}

// TLSConfig contains TLS/HTTPS related configuration options.
type TLSConfig struct {
	// Enabled enables HTTPS/TLS support.
	// Default: false
	Enabled bool `mapstructure:"enabled"`

	// CertFile specifies the path to the TLS certificate file.
	// Required when TLS is enabled.
	CertFile string `mapstructure:"cert-file"`

	// KeyFile specifies the path to the TLS private key file.
	// Required when TLS is enabled.
	KeyFile string `mapstructure:"key-file"`
}

func (c *Config) Validate() {
	if c.Name == "" {
		c.Name = "GinPlus Application"
	}
	if c.Port <= 0 || c.Port > 65535 {
		c.Port = 4006
	}
	if c.Host == "" {
		c.Host = "0.0.0.0"
	}
	if c.ContextPath == "" {
		c.ContextPath = "/"
	}
	if c.Mode == "" {
		c.Mode = "debug"
	} else if c.Mode != "debug" && c.Mode != "release" && c.Mode != "test" {
		c.Mode = "debug"
	}
	if c.MaxMultipartMemory <= 0 {
		c.MaxMultipartMemory = 8388608 // 8MB
	}
	if c.ExitDelay <= 0 {
		c.ExitDelay = 3 * time.Second // Default exit delay
	}
}
