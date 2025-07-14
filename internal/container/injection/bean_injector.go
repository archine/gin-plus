package injection

import (
	"fmt"
	"github.com/archine/gin-plus/v4/internal/container/registry"
	"reflect"
)

const (
	AutowireTag = "autowire" // tag for autowiring fields
)

// InjectBean injects a bean into a field based on the autowire field definition.
func InjectBean(bean any, fieldValue reflect.Value, autoField *registry.AutowireField) error {
	if autoField.IsInterface {
		if !reflect.TypeOf(bean).Implements(autoField.Field.Type) {
			return fmt.Errorf("bean '%s' does not implement field '%s' (type: %s)",
				autoField.AutowireTag, autoField.Name, autoField.Field.Type.String())
		}
		setFieldValue(fieldValue, bean)
		return nil
	}

	if !reflect.TypeOf(bean).AssignableTo(autoField.Field.Type) {
		return fmt.Errorf("bean '%s' (type: %s) cannot be assigned to field '%s' (type: %s)",
			autoField.AutowireTag, reflect.TypeOf(bean).String(), autoField.Name, autoField.Field.Type.String())
	}

	setFieldValue(fieldValue, bean)
	return nil
}
