package ioc

import (
	"github.com/archine/gin-plus/v4/exception"
	"github.com/archine/gin-plus/v4/internal/container"
	"reflect"
)

// RegisterBeanDefinition registers bean definitions with the IOC container
// It accepts multiple struct pointer types and registers them for dependency injection
//
// IMPORTANT: In most cases, you do NOT need to call this method manually.
// The framework will automatically discover and register beans in your project.
//
// This method is ONLY required when:
//   - Importing external dependency packages that provide interface-based injection
//   - The concrete implementations of those interfaces are defined in external packages
//   - You need to pre-register these implementations for proper dependency resolution
//
// Args:
//   - structPtrTypes: one or more struct pointer types that implement the Bean interface
//
// Example scenario:
//
//	// External package defines interface and implementation
//	// package external
//	type Logger interface {
//	    Log(msg string)
//	}
//
//	type FileLogger struct {
//	    bean.Component
//	}
//	func (f *FileLogger) Log(msg string) { /* implementation */ }
//
//	// In your main app, register the external implementation
//	RegisterBeanDefinition(reflect.TypeOf((*external.FileLogger)(nil)))
//
//	// Now you can inject the interface in your beans
//	type MyService struct {
//	    bean.Component
//	    Logger external.Logger `inject:""`
//	}
func RegisterBeanDefinition(structPtrTypes ...reflect.Type) {
	if len(structPtrTypes) == 0 {
		return
	}
	err := container.RegisterBeanDefinition(structPtrTypes)
	if err != nil {
		panic(exception.NewStackErr("failed to register bean definitions: " + err.Error()).ToString())
	}
}

// DirectSetBean allows you to directly set a bean instance in the IOC container
// This is useful for manually registering beans that are not automatically discovered
func DirectSetBean(name string, bean any) {
	if err := container.DirectSetBean(name, bean); err != nil {
		panic(exception.NewStackErr("failed to set bean: " + err.Error()).ToString())
	}
}
