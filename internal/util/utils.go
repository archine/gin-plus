package util

import (
	"crypto/md5"
	"fmt"
	"reflect"
)

// GetTypeKey generates a unique key for a given type based on its package path and name.
func GetTypeKey(typ reflect.Type) string {
	fullName := typ.PkgPath() + "." + typ.Name()
	hash := md5.Sum([]byte(fullName))
	return fmt.Sprintf("%x", hash)
}

// FirstToLower converts the first character of a string to lowercase.
func FirstToLower(s string) string {
	if len(s) == 0 {
		return s
	}
	return string(s[0]|32) + s[1:]
}
