package config

import (
	"fmt"
	"log/slog"
	"os"
	"strings"
	"time"

	"github.com/go-viper/mapstructure/v2"
	"github.com/spf13/viper"
)

const (
	// ConfigFileEnv is the environment variable holding the configuration file path.
	ConfigFileEnv = "GIN_PLUS_CONFIG_FILE"
	// DefaultConfigFile is used when ConfigFileEnv is not set.
	DefaultConfigFile = "app.yml"
)

// decodeOption lets Unmarshal honor both mapstructure and json struct tags.
var decodeOption viper.DecoderConfigOption = func(dc *mapstructure.DecoderConfig) {
	dc.TagName = "mapstructure,json"
}

// FileProvider is a Viper-based configuration implementation for loading local configuration files.
// Keys can be overridden by environment variables: the key upper-cased with '.' and '-'
// replaced by '_', e.g. gin-plus.server.context-path -> GIN_PLUS_SERVER_CONTEXT_PATH.
// Overrides apply to the Get* methods and Unmarshal("", obj), see Unmarshal.
type FileProvider struct {
	v *viper.Viper
}

// NewFileProvider creates a file-based configuration provider from the file named by the
// GIN_PLUS_CONFIG_FILE environment variable, or 'app.yml' in the working directory if unset.
//
// It panics if the file cannot be loaded, since it runs before the logger is available.
func NewFileProvider() Provider {
	path := os.Getenv(ConfigFileEnv)
	if path == "" {
		path = DefaultConfigFile
	}

	p, err := NewFileProviderAt(path)
	if err != nil {
		panic(err.Error())
	}
	return p
}

// NewFileProviderAt creates a file-based configuration provider from the given path.
// The format is inferred from the file extension (yml, yaml, json, toml, ...).
func NewFileProviderAt(path string) (Provider, error) {
	// BindStruct makes Unmarshal also look up env overrides for keys absent from the file.
	v := viper.NewWithOptions(viper.ExperimentalBindStruct())
	v.SetEnvKeyReplacer(strings.NewReplacer(".", "_", "-", "_"))
	v.AutomaticEnv()
	v.SetConfigFile(path)

	if err := v.ReadInConfig(); err != nil {
		return nil, fmt.Errorf("failed to read configuration file [%s]: %w", path, err)
	}

	// The application logger is built from this configuration, so the standard logger is used here.
	slog.Info("Successfully loaded configuration", "file", path)
	return &FileProvider{v: v}, nil
}

func (f *FileProvider) Get(key string) any {
	return f.v.Get(key)
}

func (f *FileProvider) GetBool(key string) bool {
	return f.v.GetBool(key)
}

func (f *FileProvider) GetInt(key string) int {
	return f.v.GetInt(key)
}

func (f *FileProvider) GetInt64(key string) int64 {
	return f.v.GetInt64(key)
}

func (f *FileProvider) GetInt32(key string) int32 {
	return f.v.GetInt32(key)
}

func (f *FileProvider) GetIntSlice(key string) []int {
	return f.v.GetIntSlice(key)
}

func (f *FileProvider) GetString(key string) string {
	return f.v.GetString(key)
}

func (f *FileProvider) GetStringSlice(key string) []string {
	return f.v.GetStringSlice(key)
}

func (f *FileProvider) GetStringMap(key string) map[string]any {
	return f.v.GetStringMap(key)
}

func (f *FileProvider) GetStringMapString(key string) map[string]string {
	return f.v.GetStringMapString(key)
}

func (f *FileProvider) GetStringMapStringSlice(key string) map[string][]string {
	return f.v.GetStringMapStringSlice(key)
}

func (f *FileProvider) GetUint16(key string) uint16 {
	return f.v.GetUint16(key)
}

func (f *FileProvider) GetUint32(key string) uint32 {
	return f.v.GetUint32(key)
}

func (f *FileProvider) GetUint64(key string) uint64 {
	return f.v.GetUint64(key)
}

func (f *FileProvider) GetUint(key string) uint {
	return f.v.GetUint(key)
}

func (f *FileProvider) GetDuration(key string) time.Duration {
	return f.v.GetDuration(key)
}

func (f *FileProvider) GetTime(key string) time.Time {
	return f.v.GetTime(key)
}

func (f *FileProvider) GetFloat64(key string) float64 {
	return f.v.GetFloat64(key)
}

// Sub returns a sub-configuration for the given key.
// Note: environment overrides do not apply to the returned Provider.
func (f *FileProvider) Sub(key string) Provider {
	sub := f.v.Sub(key)
	if sub == nil {
		return nil
	}
	return &FileProvider{v: sub}
}

// Unmarshal decodes the value at key (the whole configuration if key is empty) into obj.
// Note: with a non-empty key the file content is decoded as is, environment overrides of
// its nested keys do not apply (a viper limitation), use the Get* methods for those.
func (f *FileProvider) Unmarshal(key string, obj any) error {
	if key == "" {
		return f.v.Unmarshal(obj, decodeOption)
	}
	return f.v.UnmarshalKey(key, obj, decodeOption)
}
