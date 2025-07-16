package gpconf

import (
	"flag"
	"fmt"
	"os"
	"time"

	"github.com/spf13/viper"
)

// LocalFileConfigure is a Viper-based configuration implementation for loading local configuration files.
// It follows these precedence rules when determining the configuration file path:
// 1. Checks the GIN_PLUS_CONFIG_FILE environment variable first
// 2. If not set, checks for the '-c' command-line flag
// 3. Defaults to 'app.yml' in the current directory if neither is specified
type LocalFileConfigure struct {
	v *viper.Viper
}

func NewLocalFileConfigure() Configure {
	configFile := os.Getenv("GIN_PLUS_CONFIG_FILE") // Check environment variable first
	if configFile == "" {
		flag.StringVar(&configFile, "c", "app.yml", "sets the configuration file path, default app.yml")
		flag.Parse()
	}

	v := viper.New()
	v.AutomaticEnv()
	v.SetConfigFile(configFile)

	if err := v.ReadInConfig(); err != nil {
		panic(fmt.Sprintf("%s   Failed to read configuration file: %v, please check the file path: %s", time.Now().Format("2006-01-02 15:04:05"), err, configFile))
	}

	lc := &LocalFileConfigure{v: v}
	fmt.Printf("%s   Successfully loaded configuration from file: [%s]\n", time.Now().Format("2006-01-02 15:04:05"), configFile)

	return lc
}

func (l *LocalFileConfigure) Get(key string) any {
	return l.v.Get(key)
}

func (l *LocalFileConfigure) GetBool(key string) bool {
	return l.v.GetBool(key)
}

func (l *LocalFileConfigure) GetInt(key string) int {
	return l.v.GetInt(key)
}

func (l *LocalFileConfigure) GetInt64(key string) int64 {
	return l.v.GetInt64(key)
}

func (l *LocalFileConfigure) GetInt32(key string) int32 {
	return l.v.GetInt32(key)
}

func (l *LocalFileConfigure) GetIntSlice(key string) []int {
	return l.v.GetIntSlice(key)
}

func (l *LocalFileConfigure) GetString(key string) string {
	return l.v.GetString(key)
}

func (l *LocalFileConfigure) GetStringSlice(key string) []string {
	return l.v.GetStringSlice(key)
}

func (l *LocalFileConfigure) GetStringMap(key string) map[string]any {
	return l.v.GetStringMap(key)
}

func (l *LocalFileConfigure) GetStringMapString(key string) map[string]string {
	return l.v.GetStringMapString(key)
}

func (l *LocalFileConfigure) GetStringMapStringSlice(key string) map[string][]string {
	return l.v.GetStringMapStringSlice(key)
}

func (l *LocalFileConfigure) GetUint16(key string) uint16 {
	return l.v.GetUint16(key)
}

func (l *LocalFileConfigure) GetUint32(key string) uint32 {
	return l.v.GetUint32(key)
}

func (l *LocalFileConfigure) GetUint64(key string) uint64 {
	return l.v.GetUint64(key)
}

func (l *LocalFileConfigure) GetUint(key string) uint {
	return l.v.GetUint(key)
}

func (l *LocalFileConfigure) GetDuration(key string) time.Duration {
	return l.v.GetDuration(key)
}

func (l *LocalFileConfigure) GetTime(key string) time.Time {
	return l.v.GetTime(key)
}

func (l *LocalFileConfigure) GetFloat64(key string) float64 {
	return l.v.GetFloat64(key)
}

func (l *LocalFileConfigure) Sub(key string) Configure {
	sub := l.v.Sub(key)
	if sub == nil {
		return nil
	}
	return &LocalFileConfigure{v: sub}
}

func (l *LocalFileConfigure) Unmarshal(key string, obj any) error {
	if key == "" {
		return l.v.Unmarshal(obj)
	}
	return l.v.UnmarshalKey(key, obj)
}
