package application

import (
	"flag"
	"github.com/archine/gin-plus/v3/plugin/logger"
	ioc "github.com/archine/ioc"
	"github.com/spf13/viper"
	"time"
)

// Project global configuration

var Conf = &config{}

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
func LoadApplicationConfigFile(options []viper.Option) {
	var configFile string
	flag.StringVar(&configFile, "c", "app.yml", "Absolute path to the project configuration file, default app.yml")
	flag.Parse()
	confReader := viper.NewWithOptions(options...)
	confReader.SetConfigFile(configFile)
	confReader.SetDefault("server.port", 4006)
	confReader.SetDefault("server.env", Dev)
	confReader.SetDefault("server.max_file_size", 104857600)
	confReader.SetDefault("server.read_timeout", 0)  // 0 means no timeout
	confReader.SetDefault("server.write_timeout", 0) // 0 means no timeout
	confReader.AutomaticEnv()
	if err := confReader.ReadInConfig(); err != nil {
		logger.Log.Fatalf("Init project config error, %s", err.Error())
	}
	if err := confReader.Unmarshal(Conf); err != nil {
		logger.Log.Fatalf("Parse project config error, %s", err.Error())
	}
	ioc.SetBeans(confReader)
}

// GetConfReader Get config reader of the application
func GetConfReader() *viper.Viper {
	return ioc.GetBeanByName("viper.Viper").(*viper.Viper)
}
