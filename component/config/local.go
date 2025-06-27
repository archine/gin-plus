package config

import (
	"errors"
	"flag"
	"github.com/archine/gin-plus/v4/component/ioc"
	"github.com/spf13/viper"
)

// LocalFileConfigure provides a configuration implementation using Viper for local file-based configuration.
// It wraps the Viper instance and implements the Configure interface to provide type-safe
// access to configuration values from local configuration files.
type LocalFileConfigure struct {
	v *viper.Viper
}

func NewLocalFileConfigure() Configure {
	var configFile string
	flag.StringVar(&configFile, "c", "app.yml", "sets the configuration file path, default app.yml")
	flag.Parse()

	v := viper.New()
	v.AutomaticEnv()
	v.SetConfigFile(configFile)

	if err := v.ReadInConfig(); err != nil {
		panic("Failed to read the configuration file: " + err.Error())
	}

	lc := &LocalFileConfigure{v: v}
	ioc.DirectSetBean("localFileConfigure", lc)

	return lc
}

func (l *LocalFileConfigure) GetBoolean(key string) (bool, error) {
	if !l.v.IsSet(key) {
		return false, errors.New("config key not found: " + key)
	}
	return l.v.GetBool(key), nil
}

func (l *LocalFileConfigure) GetBooleanOrDefault(key string, defaultValue bool) bool {
	if !l.v.IsSet(key) {
		return defaultValue
	}
	return l.v.GetBool(key)
}

func (l *LocalFileConfigure) GetInt64(key string) (int64, error) {
	if !l.v.IsSet(key) {
		return 0, errors.New("config key not found: " + key)
	}
	return l.v.GetInt64(key), nil
}

func (l *LocalFileConfigure) GetInt(key string) (int, error) {
	if !l.v.IsSet(key) {
		return 0, errors.New("config key not found: " + key)
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

func (l *LocalFileConfigure) Sub(key string) (Configure, error) {
	if !l.v.IsSet(key) {
		return nil, errors.New("config key not found: " + key)
	}
	sub := l.v.Sub(key)
	if sub == nil {
		return nil, errors.New("config key is not a sub-configuration: " + key)
	}
	return &LocalFileConfigure{v: sub}, nil
}

func (l *LocalFileConfigure) Unmarshal(key string, obj any) error {
	if key == "" {
		return l.v.Unmarshal(obj)
	}
	if !l.v.IsSet(key) {
		return errors.New("config key not found: " + key)
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
