package injector

import (
	"reflect"

	"github.com/archine/gin-plus/v4/internal/container/util"
)

const (
	AutowireTag = "autowire" // tag for autowiring fields
)

// WireBean injects a bean into a struct field based on the provided field value.
func WireBean(bean any, fieldValue reflect.Value) error {
	return util.SetFieldValue(fieldValue, bean)
}
