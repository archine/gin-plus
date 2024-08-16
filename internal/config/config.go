package config

import (
	"flag"
	"fmt"
	"github.com/archine/gin-plus/v3/internal"
	"github.com/archine/gin-plus/v3/listener"
	"github.com/archine/ioc"
	"github.com/spf13/viper"
	"time"
)

// Project global configuration

// Conf project basic configuration
var Conf *Config

const (
	// Dev development environment
	Dev = "dev"
	// Prod production environment
	Prod = "prod"
)

type Config struct {
	Server struct {
		// Application running port, default 4006
		Port int `mapstructure:"port"`

		// Application environment, default dev, you can set it to prod or test
		Env string `mapstructure:"env"`

		// AllowedCors allowed cross-domain, default false
		// If true, the server will add the default cors middleware to the gin engine.
		//
		// Note: When you add cross-domain middleware through application.New(), you do not need to allow it, otherwise there will be multiple,
		// this parameter only controls when you use application.Default()
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

		// DisableGeneralOptions Disable general options, default true.
		// if true, passes "OPTIONS *" requests to the Handler.
		// otherwise responds with 200 OK and Content-Length: 0.
		DisableGeneralOptions bool `mapstructure:"disable_general_options"`
	}
}

// LoadByCommand Load configuration file by command line
func LoadByCommand(l listener.ConfigListener) {
	var configFile string
	flag.StringVar(&configFile, "c", "app.yml", "Absolute path to the project configuration file, default app.yml")
	flag.Parse()
	Load(configFile, l)
}

// Load Load configuration file
func Load(configFilePath string, l listener.ConfigListener) {
	v := viper.New()
	ioc.SetBeans(v)

	v.SetDefault("server.port", 4006)
	v.SetDefault("server.env", Dev)
	v.SetDefault("server.max_multipart_memory", 32<<20) // 32M
	v.SetDefault("server.read_timeout", 0)
	v.SetDefault("server.write_timeout", 0)
	v.SetDefault("server.read_header_timeout", 0)
	v.SetDefault("server.idle_timeout", 0)
	v.SetDefault("server.disable_general_options", true)
	v.AutomaticEnv()
	v.SetConfigFile(configFilePath)

	var err error
	if l != nil {
		err = l.Read(v)
	} else {
		err = v.ReadInConfig()
	}
	if err != nil {
		panic(fmt.Sprintf("Error reading the configuration file: %s", err.Error()))
	}

	err = v.Unmarshal(&Conf)
	if err != nil {
		panic(fmt.Sprintf("Error unmarshalling the configuration file into the config structure: %s", err.Error()))
	}

	if l != nil {
		l.After(v)
	}
	internal.Logger.Info("Configuration file loaded successfully.")
}
