package application

import (
	"github.com/archine/gin-plus/v3/internal/config"
	"github.com/archine/gin-plus/v3/ioc"
	"github.com/spf13/viper"
)

// GetConfReader returns the configuration reader of the application.
func GetConfReader() *viper.Viper {
	return ioc.GetBeanByName("viper.Viper").(*viper.Viper)
}

// GetConfig returns the basic configuration of the application.
func GetConfig() *config.Config {
	return config.Conf
}
