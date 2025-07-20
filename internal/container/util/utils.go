package util

import (
	"errors"
	"reflect"
	"unsafe"
)

// SetFieldValue sets a struct field to the given value, supporting both direct assignment and type conversion.
// If the field cannot be set directly, it uses unsafe pointers to assign the value.
func SetFieldValue(fieldValue reflect.Value, value any) error {
	if value == nil {
		return nil
	}

	reflectValue := reflect.ValueOf(value)
	valueType := reflectValue.Type()
	fieldType := fieldValue.Type()

	if valueType == fieldType {
		setValue(fieldValue, reflectValue)
		return nil
	}

	if valueType.AssignableTo(fieldType) {
		setValue(fieldValue, reflectValue)
	} else if valueType.ConvertibleTo(fieldType) {
		setValue(fieldValue, reflectValue.Convert(fieldType))
	} else {
		return errors.New("type mismatch")
	}

	return nil
}

func setValue(fieldValue reflect.Value, value reflect.Value) {
	if fieldValue.CanSet() {
		fieldValue.Set(value)
	} else {
		ptr := unsafe.Pointer(fieldValue.UnsafeAddr())
		newValue := reflect.NewAt(fieldValue.Type(), ptr).Elem()
		newValue.Set(value)
	}
}
