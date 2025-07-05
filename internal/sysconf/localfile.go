package sysconf

import (
	"flag"
	"fmt"
	"github.com/archine/gin-plus/v4/component/ioc"
	"github.com/archine/gin-plus/v4/util/reflectutil"

	"github.com/archine/gin-plus/v4/component/config"
	"github.com/spf13/viper"
)

// LocalFileConfigure provides a configuration implementation using Viper for local file-based configuration.
// It wraps the Viper instance and implements the Configure interface to provide type-safe
// access to configuration values from local configuration files.
type LocalFileConfigure struct {
	v *viper.Viper
}

func NewLocalFileConfigure() config.Configure {
	var configFile string
	flag.StringVar(&configFile, "c", "app.yml", "sets the configuration file path, default app.yml")
	flag.Parse()

	v := viper.New()
	v.AutomaticEnv()
	v.SetConfigFile(configFile)

	if err := v.ReadInConfig(); err != nil {
		panic(fmt.Sprintf("Failed to load configuration file '%s': %v\nPlease ensure the file exists and is properly formatted", configFile, err))
	}

	lc := &LocalFileConfigure{v: v}

	err := ioc.RegisterBean("", lc, reflectutil.InterfaceOf[config.Configure]())
	if err != nil {
		panic(fmt.Sprintf("Failed to register LocalFileConfigure in IoC container: %v", err))
	}

	fmt.Printf("Configuration loaded successfully from: %s\n", configFile)
	return lc
}

func (l *LocalFileConfigure) GetBool(key string) (bool, error) {
	if !l.v.IsSet(key) {
		return false, fmt.Errorf("config key '%s' not found", key)
	}
	return l.v.GetBool(key), nil
}

func (l *LocalFileConfigure) GetBoolOrDefault(key string, defaultValue bool) bool {
	if !l.v.IsSet(key) {
		return defaultValue
	}
	return l.v.GetBool(key)
}

func (l *LocalFileConfigure) GetInt64(key string) (int64, error) {
	if !l.v.IsSet(key) {
		return 0, fmt.Errorf("config key '%s' not found", key)
	}
	return l.v.GetInt64(key), nil
}

func (l *LocalFileConfigure) GetInt(key string) (int, error) {
	if !l.v.IsSet(key) {
		return 0, fmt.Errorf("config key '%s' not found", key)
	}
	return l.v.GetInt(key), nil
}

func (l *LocalFileConfigure) GetIntOrDefault(key string, defaultValue int) int {
	if !l.v.IsSet(key) {
		return defaultValue
	}
	return l.v.GetInt(key)
}

func (l *LocalFileConfigure) GetString(key string) string {
	return l.v.GetString(key)
}

func (l *LocalFileConfigure) GetStringOrDefault(key string, defaultValue string) string {
	if !l.v.IsSet(key) {
		return defaultValue
	}
	return l.v.GetString(key)
}

func (l *LocalFileConfigure) Sub(key string) (config.Configure, error) {
	if !l.v.IsSet(key) {
		return nil, fmt.Errorf("config key '%s' not found", key)
	}
	sub := l.v.Sub(key)
	if sub == nil {
		return nil, fmt.Errorf("config key '%s' is not a valid sub-configuration", key)
	}
	return &LocalFileConfigure{v: sub}, nil
}

func (l *LocalFileConfigure) Unmarshal(key string, obj any) error {
	if key == "" {
		return l.v.Unmarshal(obj)
	}
	if !l.v.IsSet(key) {
		return fmt.Errorf("config key '%s' not found", key)
	}
	return l.v.UnmarshalKey(key, obj)
}

func (l *LocalFileConfigure) GetStringSlice(key string) []string {
	return l.v.GetStringSlice(key)
}

func (l *LocalFileConfigure) Get(key string) any {
	if !l.v.IsSet(key) {
		return nil
	}
	return l.v.Get(key)
}
