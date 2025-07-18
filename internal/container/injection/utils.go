package injection

import (
	"reflect"
	"unsafe"
)

// setFieldValue sets a struct field to the given value, supporting both direct assignment and type conversion.
// If the field cannot be set directly, it uses unsafe pointers to assign the value.
// Commonly used for autowiring in dependency injection scenarios.
// Does nothing if the provided value is nil.
func setFieldValue(fieldValue reflect.Value, value any) {
	if value == nil {
		return
	}

	reflectValue := reflect.ValueOf(value)
	valueType := reflectValue.Type()
	fieldType := fieldValue.Type()

	if valueType == fieldType {
		setValue(fieldValue, reflectValue)
		return
	}

	if valueType.AssignableTo(fieldType) {
		setValue(fieldValue, reflectValue)
	} else if valueType.ConvertibleTo(fieldType) {
		setValue(fieldValue, reflectValue.Convert(fieldType))
	}
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
