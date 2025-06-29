package config

// Configure defines the interface for configuration management operations
type Configure interface {
	// GetBool retrieves a boolean value for the given key, returns error if key not found
	GetBool(s string) (bool, error)

	// GetBoolOrDefault retrieves a boolean value for the given key, returns defaultValue if key not found
	GetBoolOrDefault(s string, defaultValue bool) bool

	// GetInt64 retrieves an int64 value for the given key, returns error if key not found
	GetInt64(s string) (int64, error)

	// GetInt retrieves an int value for the given key, returns error if key not found
	GetInt(s string) (int, error)

	// GetIntOrDefault retrieves an int value for the given key, returns defaultValue if key not found
	GetIntOrDefault(s string, defaultValue int) int

	// GetString retrieves a string value for the given key, returns empty string if key not found
	GetString(s string) string

	// GetStringOrDefault retrieves a string value for the given key, returns defaultValue if key not found
	GetStringOrDefault(s string, defaultValue string) string

	// Sub returns a sub-configuration for the given key
	Sub(s string) (Configure, error)

	// Unmarshal the configuration into the provided object
	Unmarshal(s string, obj interface{}) error

	// GetStringSlice retrieves a string slice for the given key
	GetStringSlice(s string) []string

	// Get retrieves a value of any type for the given key
	Get(key string) any
}
