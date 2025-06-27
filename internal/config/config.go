package config

import (
	"time"
)

type Config struct {
	Server struct {
		// Application running port, default 4006
		Port int `mapstructure:"port"`

		// Application environment, default dev, you can set it to prod
		Env string `mapstructure:"env"`

		// AllowedCors allowed cross-domain, default false
		// If true, the server will add the default cors middleware to the gin engine.
		AllowedCors bool `mapstructure:"allowed_cors"`

		// MaxMultipartMemory The maximum memory space that an uploaded file can occupy, default 32M.
		// When the uploaded file exceeds this limit, Gin writes the file data to a temporary file instead of storing it directly in memory.
		// This can help avoid the problem of excessive memory consumption caused by large files.
		MaxMultipartMemory int64 `mapstructure:"max_multipart_memory"`

		// WriteTimeout is the maximum duration before timing out
		// writes of the response. It is reset whenever a new
		// request's header is read. Like ReadTimeout, it does not
		// let Handlers make decisions on a per-request basis.
		// A zero or negative value means there will be no timeout.
		WriteTimeout time.Duration `mapstructure:"write_timeout"`

		// ReadTimeout is the maximum duration for reading the entire
		// request, including the body. A zero or negative value means
		// there will be no timeout.
		//
		// Because ReadTimeout does not let Handlers make per-request
		// decisions on each request body's acceptable deadline or
		// upload rate, most users will prefer to use
		// ReadHeaderTimeout. It is valid to use them both.
		ReadTimeout time.Duration `mapstructure:"read_timeout"`

		// ReadHeaderTimeout is the amount of time allowed to read
		// request headers. The connection's read deadline is reset
		// after reading the headers and the Handler can decide what
		// is considered too slow for the body. If ReadHeaderTimeout
		// is zero, the value of ReadTimeout is used. If both are
		// zero, there is no timeout.
		ReadHeaderTimeout time.Duration `mapstructure:"read_header_timeout"`

		// IdleTimeout is the maximum amount of time to wait for the
		// next request when keep-alives are enabled. If IdleTimeout
		// is zero, the value of ReadTimeout is used. If both are
		// zero, there is no timeout.
		IdleTimeout time.Duration `mapstructure:"idle_timeout"`

		// ShutdownTimeout specifies the maximum duration to wait before the server
		// gracefully shuts down. If the value is zero, the server will not have a timeout
		// and will wait indefinitely for ongoing requests to complete.
		ShutdownTimeout time.Duration `mapstructure:"shutdown_timeout"`

		// DisableGeneralOptions Disable general options, default true.
		// if true, passes "OPTIONS *" requests to the Handler.
		// otherwise responds with 200 OK and Content-Length: 0.
		DisableGeneralOptions bool `mapstructure:"disable_general_options"`
	}
}
