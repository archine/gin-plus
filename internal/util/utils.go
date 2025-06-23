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
