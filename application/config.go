package application

import (
	"flag"
	"fmt"
	"github.com/archine/gin-plus/v3/listener"
	ioc "github.com/archine/ioc"
	"github.com/spf13/viper"
	"time"
)

// Project global configuration

// Conf project basic configuration
var Conf *config

const (
	// Dev development environment
	Dev = "dev"
	// Test environment
	Test = "test"
	// Prod production environment
	Prod = "prod"
)

type config struct {
	Server struct {
		Port         int           `mapstructure:"port"`          // Application port
		Env          string        `mapstructure:"env"`           // Application environment, default dev, you can set it to prod or test
		MaxFileSize  int64         `mapstructure:"max_file_size"` // Maximum file size, default 100M
		WriteTimeout time.Duration `mapstructure:"write_timeout"` // Write timeout, default 0 means no timeout
		ReadTimeout  time.Duration `mapstructure:"read_timeout"`  // Read timeout, default 0 means no timeout
	}
}

// LoadApplicationConfigFile load the application configuration file
func LoadApplicationConfigFile(l listener.ConfigListener) {
	var v = viper.New()
	var configFile string
	flag.StringVar(&configFile, "c", "app.yml", "Absolute path to the project configuration file, default app.yml")
	flag.Parse()
	v.SetConfigFile(configFile)
	v.SetDefault("server.port", 4006)
	v.SetDefault("server.env", Dev)
	v.SetDefault("server.max_file_size", 104857600)
	v.SetDefault("server.read_timeout", 0)  // 0 means no timeout
	v.SetDefault("server.write_timeout", 0) // 0 means no timeout
	v.AutomaticEnv()
	var err error
	if l != nil {
		err = l.Read(v)
	} else {
		err = v.ReadInConfig()
	}
	if err != nil {
		panic(fmt.Sprintf("Failed to read the configuration file, %s", err.Error()))
	}
	if err = v.Unmarshal(&Conf); err != nil {
		panic(fmt.Sprintf("Failed to parse the configuration file, %s", err.Error()))
	}
	ioc.SetBeans(v)
}

// GetConfReader Get config reader of the application
func GetConfReader() *viper.Viper {
	return ioc.GetBeanByName("viper.Viper").(*viper.Viper)
}
