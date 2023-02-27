package application

import (
	"flag"
	log "github.com/sirupsen/logrus"
	"github.com/spf13/viper"
)

// Project global configuration

var Env = &config{}

// LoadApplicationConfigFile load the application configuration file
func LoadApplicationConfigFile() {
	var configFile string
	flag.StringVar(&configFile, "c", "app.yml", "project application path, default app.yml")
	flag.Parse()
	viper.SetDefault("log_level", "debug")
	viper.SetDefault("port", 4006)
	viper.SetDefault("max_file_zie", 104857600)
	viper.SetConfigFile(configFile)
	viper.AutomaticEnv()
	if err := viper.ReadInConfig(); err != nil {
		log.Fatalf("Init project config error, %s", err.Error())
	}
	if err := viper.Unmarshal(Env); err != nil {
		log.Fatalf("Parse project config error, %s", err.Error())
	}
}

type config struct {
	LogLevel    string `mapstructure:"log_level"`     // Project global log level, support error, info, trace, warn, panic, fetal, and debug
	Port        int    `mapstructure:"port"`          // Project run port
	MaxFileSize int64  `mapstructure:"max_file_size"` // Maximum file upload size，Unit byte
}
