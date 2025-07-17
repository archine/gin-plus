package injection

import (
	"reflect"
	"unsafe"
)

// setFieldValue sets the value of a field in a struct, handling both settable and unsafe cases.
func setFieldValue(fieldValue reflect.Value, value any) {
	var reflectValue reflect.Value
	if value == nil {
		reflectValue = reflect.Zero(fieldValue.Type())
	} else {
		reflectValue = reflect.ValueOf(value)
	}

	if fieldValue.CanSet() {
		if reflectValue.Type().AssignableTo(fieldValue.Type()) {
			fieldValue.Set(reflectValue)
		} else if reflectValue.Type().ConvertibleTo(fieldValue.Type()) {
			fieldValue.Set(reflectValue.Convert(fieldValue.Type()))
		}
	} else {
		ptr := unsafe.Pointer(fieldValue.UnsafeAddr())
		newValue := reflect.NewAt(fieldValue.Type(), ptr).Elem()

		if reflectValue.Type().AssignableTo(fieldValue.Type()) {
			newValue.Set(reflectValue)
		} else if reflectValue.Type().ConvertibleTo(fieldValue.Type()) {
			newValue.Set(reflectValue.Convert(fieldValue.Type()))
		}
	}
}
