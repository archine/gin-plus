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
func CopyProperties(src interface{}, target interface{}) error {
	srcProxy := reflect.ValueOf(src)
	srcTypeof := reflect.TypeOf(src)
	if srcProxy.Kind() == reflect.Pointer {
		srcProxy = srcProxy.Elem()
		srcTypeof = srcTypeof.Elem()
	}
	if srcProxy.Kind() != reflect.Struct {
		return errors.New("src must be a struct")
	}
	targetProxy := reflect.ValueOf(target).Elem()
	if targetProxy.Kind() != reflect.Struct && targetProxy.Kind() != reflect.Pointer {
		return errors.New("target must be a struct pointer")
	}
	var filedName string
	for i := 0; i < srcTypeof.NumField(); i++ {
		f := srcTypeof.Field(i)
		filedName = f.Tag.Get("copy")
		switch filedName {
		case "-":
			continue
		case "":
			filedName = f.Name
		}
		targetField := targetProxy.FieldByName(filedName)
		if targetField.Kind() == reflect.Invalid {
			continue
		}
		targetField.Set(srcProxy.Field(i))
	}
	return nil
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
