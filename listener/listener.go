package listener

import "github.com/spf13/viper"

// ApplicationListener Application listener
type ApplicationListener interface{}

// ApplicationEventListener Application event listener
type ApplicationEventListener interface {
	ApplicationListener

	// PreApply triggered before mvc starts, Before the project starts.
	// This is where you can provide basic services, such as set beans.
	// Of course, you can also perform logic here that doesn't require obtaining beans.
	PreApply()

	// PreStart The last event before the project starts, dependency injection is all finished and ready to run.
	// You can execute any logic here.
	PreStart()

	// PreStop The event before the application stops can be performed here to close some resources
	PreStop()

	// PostStop Events after the application has stopped can perform other closing operations here
	PostStop()
}

// ConfigListener Configuration listener, used to load configuration
type ConfigListener interface {
	ApplicationListener

	// Read configuration
	Read(v *viper.Viper) error
}

// DoPreApply Trigger the PreApply event
func DoPreApply(listeners []ApplicationListener) {
	for _, l := range listeners {
		if ael, ok := l.(ApplicationEventListener); ok {
			ael.PreApply()
		}
	}
}

// DoPreStart Trigger the PreStart event
func DoPreStart(listeners []ApplicationListener) {
	for _, l := range listeners {
		if ael, ok := l.(ApplicationEventListener); ok {
			ael.PreStart()
		}
	}
}

// DoPreStop Trigger the PreStop event
func DoPreStop(listeners []ApplicationListener) {
	for _, l := range listeners {
		if ael, ok := l.(ApplicationEventListener); ok {
			ael.PreStop()
		}
	}
}

// DoPostStop Trigger the PostStop event
func DoPostStop(listeners []ApplicationListener) {
	for _, l := range listeners {
		if ael, ok := l.(ApplicationEventListener); ok {
			ael.PostStop()
		}
	}
}
