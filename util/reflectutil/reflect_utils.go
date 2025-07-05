package reflectutil

import "reflect"

// InterfaceOf returns the reflect.Type of interface T.
// This is commonly used for getting the type of an interface without creating an instance.
//
// For example, if T is an interface type MyInterface, InterfaceOf(T) returns reflect.Type representing MyInterface.
func InterfaceOf[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}

// StructOf returns the reflect.Type of struct T.
// This is useful for getting the type of a struct without creating an instance.
//
// For example, if T is a struct type Foo, StructOf(T) returns reflect.Type representing Foo.
func StructOf[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil)).Elem()
}

// PtrOf PointerOf returns the reflect.Type of pointer to T.
// This is useful when you need the pointer type for reflection operations.
//
// For example, if T is a struct type Foo, PtrOf(T) returns reflect.Type representing *Foo.
func PtrOf[T any]() reflect.Type {
	return reflect.TypeOf((*T)(nil))
}
