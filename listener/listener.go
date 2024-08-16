package listener

import "github.com/spf13/viper"

// ApplicationListener Application listener
type ApplicationListener interface{}

// ApplicationEventListener defines an interface for listening to application lifecycle events.
type ApplicationEventListener interface {
	ApplicationListener

	// PreApply is triggered before the MVC framework starts, and before the application fully initiates.
	// This is the stage where you can set up basic services, such as registering beans.
	// Additionally, any logic that doesn't require access to beans can be executed here.
	PreApply()

	// PreStart is the final event before the application starts running, after all dependency injections have been completed.
	// At this point, you can execute any necessary logic to prepare for the application's startup.
	PreStart()

	// PreStop is triggered before the application stops, providing an opportunity to close resources
	// or perform other pre-shutdown tasks.
	PreStop()

	// PostStop is triggered after the application has stopped, allowing for final cleanup operations
	// or any other shutdown-related activities.
	PostStop()
}

// ConfigListener defines an interface for configuration listeners
// that are used to load and process configuration settings.
type ConfigListener interface {
	ApplicationListener

	// Read loads the configuration settings from the provided Viper instance.
	// This method is intended to be implemented by users to define custom
	// configuration loading logic.
	Read(v *viper.Viper) error

	// After is called after the configuration has been successfully read.
	// This method allows for any additional processing or setup that may
	// be required after the initial configuration load.
	After(v *viper.Viper)
}

// DoPreApply triggers the PreApply event.
func DoPreApply(listeners []ApplicationListener) {
	triggerEvent(listeners, ApplicationEventListener.PreApply)
}

// DoPreStart triggers the PreStart event.
func DoPreStart(listeners []ApplicationListener) {
	triggerEvent(listeners, ApplicationEventListener.PreStart)
}

// DoPreStop triggers the PreStop event.
func DoPreStop(listeners []ApplicationListener) {
	triggerEvent(listeners, ApplicationEventListener.PreStop)
}

// DoPostStop triggers the PostStop event.
func DoPostStop(listeners []ApplicationListener) {
	triggerEvent(listeners, ApplicationEventListener.PostStop)
}

// triggerEvent triggers the specified event method for all listeners that implement ApplicationEventListener.
func triggerEvent(listeners []ApplicationListener, eventFunc func(ApplicationEventListener)) {
	for _, l := range listeners {
		if ael, ok := l.(ApplicationEventListener); ok {
			eventFunc(ael)
		}
	}
}
