package event

import "github.com/spf13/viper"

// ConfigBeforeLoadEvent handles events before configuration loading begins.
// Implement this interface to customize pre-loading behavior such as setting defaults or preparing config sources.
type ConfigBeforeLoadEvent interface {
	AppEvent

	// OnConfigBeforeLoad is called before configuration files are read from disk.
	// The viper instance is provided to allow pre-configuration setup like setting defaults,
	// adding config paths, or configuring environment variable mappings.
	// Use this hook to prepare the configuration system before actual loading occurs.
	OnConfigBeforeLoad(v *viper.Viper)
}

// ConfigAfterLoadEvent handles events after configuration loading completes.
// Implement this interface to perform post-loading tasks such as validation or derived value calculation.
type ConfigAfterLoadEvent interface {
	AppEvent

	// OnConfigAfterLoad is called after all configuration files have been successfully loaded.
	// The viper instance contains all loaded configuration values and can be used for
	// validation, transformation, or setting up derived configuration values.
	// Use this hook to ensure configuration integrity and perform any necessary post-processing.
	OnConfigAfterLoad(v *viper.Viper)
}
