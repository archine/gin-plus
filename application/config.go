package application

import (
	"flag"
	"github.com/archine/ioc"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Project global configuration

var Env = &config{}

type config struct {
	LogLevel    string `mapstructure:"log_level"`     // The log level of project, default debug
	Port        int    `mapstructure:"port"`          // Project run port, default 4006
	MaxFileSize int64  `mapstructure:"max_file_size"` // Maximum file upload size，default 100Mb
}

// LoadApplicationConfigFile load the application configuration file
func LoadApplicationConfigFile(options []viper.Option) *viper.Viper {
	var configFile string
	flag.StringVar(&configFile, "c", "app.yml", "Absolute path to the project configuration file, default app.yml")
	flag.Parse()
	confReader := viper.NewWithOptions(options...)
	confReader.SetConfigFile(configFile)
	confReader.SetDefault("log_level", "debug")
	confReader.SetDefault("port", 4006)
	confReader.SetDefault("max_file_zie", 104857600)
	confReader.AutomaticEnv()
	if err := confReader.ReadInConfig(); err != nil {
		log.Fatalf("Init project config error, %s", err.Error())
	}
	if err := confReader.Unmarshal(Env); err != nil {
		log.Fatalf("Parse project config error, %s", err.Error())
	}
	ioc.SetBeans(confReader)
	return confReader
}

// GetConfReader Get config reader of the application
func GetConfReader() *viper.Viper {
	return ioc.GetBeanByName("viper.Viper").(*viper.Viper)
}
