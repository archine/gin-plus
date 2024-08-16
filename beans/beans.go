package beans

import (
	"errors"
	"reflect"
)

// CopyProperties copies data from the source structure to the target structure.
//
// src: The source structure, which can be either a pointer or a value.
// target: The target structure, which must be a pointer.
//
// Example:
//
//	type Src struct {
//	    Gender   int                  // Field name is used by default if not specified
//	    Age      int    `copy:"Age"`  // Specify field name
//	    Username string `copy:"-"`    // Ignore the field by using "-", or make it private
//	}
//
//	type Target struct {
//	    Age      int
//	    Username string
//	    Gender   int
//	}
//
//	err := Beans.CopyProperties(&src, &target)
func CopyProperties(src any, target any) error {
	srcValue := reflect.Indirect(reflect.ValueOf(src))
	srcType := srcValue.Type()

	if srcValue.Kind() != reflect.Struct {
		return errors.New("src must be a struct or a pointer to a struct")
	}

	targetValue := reflect.ValueOf(target).Elem()
	if targetValue.Kind() != reflect.Struct {
		return errors.New("target must be a pointer to a struct")
	}

	copyFields(srcType, srcValue, targetValue)
	return nil
}

func copyFields(srcType reflect.Type, srcValue, targetValue reflect.Value) {
	for i := 0; i < srcType.NumField(); i++ {
		field := srcType.Field(i)
		if field.Anonymous && field.Type.Kind() == reflect.Struct {
			copyFields(field.Type, srcValue.Field(i), targetValue)
			continue
		}

		fieldName := field.Tag.Get("copy")
		if fieldName == "-" {
			continue
		}
		if fieldName == "" {
			fieldName = field.Name
		}

		targetField := targetValue.FieldByName(fieldName)
		if !targetField.IsValid() || field.Type != targetField.Type() {
			continue
		}

		targetField.Set(srcValue.Field(i))
	}
}

// ToMap converts a structure to a map.
//
// Notes: Non-exported fields are not included.
//
// Example:
//
//	type User struct {
//	    Username string             // Field name is used as the key
//	    UserAge  int `alias:"年龄"`  // Field name is replaced by an alias
//	}
//
//	beans.ToMap(&user)
func ToMap(val any) (map[string]any, error) {
	valProxy := reflect.Indirect(reflect.ValueOf(val))
	valType := valProxy.Type()

	if valProxy.Kind() != reflect.Struct {
		return nil, errors.New("val must be a struct or a pointer to a struct")
	}

	result := make(map[string]any)
	for i := 0; i < valType.NumField(); i++ {
		field := valType.Field(i)
		key := field.Tag.Get("alias")
		if key == "" {
			key = field.Name
		}
		result[key] = valProxy.Field(i).Interface()
	}
	return result, nil
}
