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
		DirectSetValue(fieldValue, reflectValue)
		return nil
	}

	if valueType.AssignableTo(fieldType) {
		DirectSetValue(fieldValue, reflectValue)
	} else if valueType.ConvertibleTo(fieldType) {
		DirectSetValue(fieldValue, reflectValue.Convert(fieldType))
	} else {
		return errors.New("type mismatch")
	}

	return nil
}

// DirectSetValue sets a value to a field directly, bypassing the normal setter checks.
func DirectSetValue(fieldValue reflect.Value, value reflect.Value) {
	if fieldValue.CanSet() {
		fieldValue.Set(value)
	} else {
		ptr := unsafe.Pointer(fieldValue.UnsafeAddr())
		newValue := reflect.NewAt(fieldValue.Type(), ptr).Elem()
		newValue.Set(value)
	}
}
