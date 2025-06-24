package container

import (
	"github.com/archine/gin-plus/v4/component/bean"
	"reflect"
)

// Definition is a structure that defines the properties of a bean.
// Create the bean according to this definition at startup
type Definition struct {
	Type reflect.Type      // Type of the bean, usually the type of the structure implementing the
	Bean bean.AbstractBean // Instance of the bean, usually the structure implementing the AbstractBean interface
}
