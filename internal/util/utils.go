package util

import (
	"fmt"
	"reflect"
)

// GetTypeKey generates a unique key for a given type based on its package path and name.
func GetTypeKey(typ reflect.Type) string {
	if typ.Kind() == reflect.Ptr {
		typ = typ.Elem()
	}
	packagePath := typ.PkgPath()
	typeName := typ.Name()
	return fmt.Sprintf("%s.%s", packagePath, typeName)
}

// FirstToLower converts the first character of a string to lowercase.
func FirstToLower(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(s[0]|32) + s[1:]
}
