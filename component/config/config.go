package config

import "time"

// Provider defines the interface for configuration management operations
type Provider interface {
	// Get retrieves a value of any type for the given key
	Get(key string) any

	// GetBool retrieves a boolean value for the given key
	GetBool(key string) bool

	// GetInt retrieves an int value for the given key
	GetInt(key string) int

	// GetInt64 retrieves an int64 value for the given key
	GetInt64(key string) int64

	// GetInt32 retrieves an int32 value for the given key
	GetInt32(key string) int32

	// GetIntSlice retrieves an int slice for the given key
	GetIntSlice(key string) []int

	// GetString retrieves a string value for the given key
	GetString(key string) string

	// GetStringSlice retrieves a string slice for the given key
	GetStringSlice(key string) []string

	// GetStringMap retrieves a string map for the given key
	GetStringMap(key string) map[string]any

	// GetStringMapString retrieves a string map with string values for the given key
	GetStringMapString(key string) map[string]string

	// GetStringMapStringSlice retrieves a string map with string slice values for the given key
	GetStringMapStringSlice(key string) map[string][]string

	// GetUint16 retrieves a uint16 value for the given key
	GetUint16(key string) uint16

	// GetUint32 retrieves a uint32 value for the given key
	GetUint32(key string) uint32

	// GetUint64 retrieves a uint64 value for the given key
	GetUint64(key string) uint64

	// GetUint retrieves a uint value for the given key
	GetUint(key string) uint

	// GetDuration retrieves a time.Duration value for the given key
	GetDuration(key string) time.Duration

	// GetTime retrieves a time.Time value for the given key
	GetTime(key string) time.Time

	// GetFloat64 retrieves a float64 value for the given key
	GetFloat64(key string) float64

	// Sub returns a sub-configuration for the given key
	Sub(key string) Provider

	// Unmarshal the configuration into the provided object
	Unmarshal(key string, obj any) error
}
