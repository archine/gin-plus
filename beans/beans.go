package beans

import (
	"errors"
	"reflect"
)

/*
CopyProperties copy the source structure data to the target structure.

	src: source structure, type can be pointer and normal.
	target: target structure, type must be a pointer.

Example:

	type Src struct {
	    Gender   int                  // The name of the field is used when not specified
	    Age      int    `copy:"Age"`  // Specify field name
	    Username string `copy:"-"`    // Ignore the copy by setting "-", or make the field private
	}

	type Target struct {
	    Age      int
	    Username string
	    Gender   int
	}

	err := Beans.CopyProperties(&src, &target)
*/
func CopyProperties(src any, target any) error {
	srcType := reflect.TypeOf(src)
	srcValue := reflect.ValueOf(src)
	if srcValue.Kind() == reflect.Pointer {
		srcValue = srcValue.Elem()
		srcType = srcType.Elem()
	}
	if srcValue.Kind() != reflect.Struct {
		return errors.New("src must be a struct")
	}
	targetValue := reflect.ValueOf(target)
	if targetValue.Kind() != reflect.Pointer {
		return errors.New("target must be a struct pointer")
	}
	targetValue = targetValue.Elem()
	if targetValue.Kind() != reflect.Struct {
		return errors.New("target must be a struct pointer")
	}
	copyCore(srcType, srcValue, targetValue)
	return nil
}

func copyCore(srcType reflect.Type, srcValue, targetValue reflect.Value) {
	var fieldName string
	for i := 0; i < srcType.NumField(); i++ {
		f := srcType.Field(i)
		if f.Anonymous && f.Type.Kind() == reflect.Struct {
			copyCore(f.Type, srcValue.Field(i), targetValue)
			continue
		}
		fieldName = f.Tag.Get("copy")
		switch fieldName {
		case "-":
			continue
		case "":
			fieldName = f.Name
		}
		targetField := targetValue.FieldByName(fieldName)
		if targetField.Kind() == reflect.Invalid {
			continue
		}
		fv := srcValue.Field(i)
		if fv.Kind() != targetField.Kind() {
			// Ignore fields of different types
			continue
		}
		targetField.Set(fv)
	}
}

/*
ToMap structure to map.

	Notes: Non-public fields are not processed

	Example:

	   type User struct {
	       Username string             // the filed name is used as the key
	       UserAge  int `alias:"年龄"`  // replace the field name with an alias
	   }

	   beans.ToMap(&user)
*/
func ToMap(val interface{}) (map[string]interface{}, error) {
	valProxy := reflect.ValueOf(val)
	valTypeof := reflect.TypeOf(val)
	if valProxy.Kind() == reflect.Pointer {
		valProxy = valProxy.Elem()
		valTypeof = valTypeof.Elem()
	}
	if valProxy.Kind() != reflect.Struct {
		return nil, errors.New("val must be a struct")
	}
	var k string
	result := make(map[string]interface{})
	for i := 0; i < valTypeof.NumField(); i++ {
		f := valTypeof.Field(i)
		k = f.Tag.Get("alias")
		if k == "" {
			k = f.Name
		}
		result[k] = valProxy.Field(i).Interface()
	}
	return result, nil
}
