package config

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

// FileProvider is a Viper-based configuration implementation for loading local configuration files.
// 1. Checks the GIN_PLUS_CONFIG_FILE environment variable first
// 2. If not set, checks for the '-c' command-line flag
// 3. Defaults to 'app.yml' in the current directory if neither is specified
type FileProvider struct {
	v *viper.Viper
}

// NewFileProvider creates a new file-based configuration provider
func NewFileProvider() Provider {
	configFile := os.Getenv("GIN_PLUS_CONFIG_FILE") // Check environment variable first
	if configFile == "" {
		flag.StringVar(&configFile, "c", "app.yml", "sets the configuration file path, default app.yml")
		flag.Parse()
	}

	v := viper.New()
	v.AutomaticEnv()
	v.SetConfigFile(configFile)

	if err := v.ReadInConfig(); err != nil {
		panic(fmt.Sprintf("failed to read configuration file: %s, %s", configFile, err.Error()))
	}

	lc := &FileProvider{v: v}
	_, _ = fmt.Fprintf(os.Stderr, "%s  Successfully loaded configuration from file: [%s]\n",
		time.Now().Format("2006-01-02 15:04:05"), configFile)

	return lc
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

func (f *FileProvider) Sub(key string) Provider {
	sub := f.v.Sub(key)
	if sub == nil {
		return nil
	}
	return &FileProvider{v: sub}
}

func (f *FileProvider) Unmarshal(key string, obj any) error {
	if key == "" {
		return f.v.Unmarshal(obj)
	}
	return f.v.UnmarshalKey(key, obj)
}
