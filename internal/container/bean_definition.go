package container

import (
	"reflect"
)

// Definition is a structure that defines the properties of a bean.
// Create the bean according to this definition at startup
type Definition struct {
	Type reflect.Type // Type of the bean, usually the type of the structure implementing the
	Bean any          // Instance of the bean, usually the structure implementing the AbstractBean interface
}
