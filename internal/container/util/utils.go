package util

import (
	"reflect"
	"unsafe"
)

// DirectSetValue sets a value to a field directly, bypassing exported-field checks via unsafe.
// fieldValue must be addressable when it is not settable.
func DirectSetValue(fieldValue reflect.Value, value reflect.Value) {
	if fieldValue.CanSet() {
		fieldValue.Set(value)
		return
	}
	if !fieldValue.CanAddr() {
		panic("DirectSetValue: field is neither settable nor addressable")
	}
	reflect.NewAt(fieldValue.Type(), unsafe.Pointer(fieldValue.UnsafeAddr())).Elem().Set(value)
}
