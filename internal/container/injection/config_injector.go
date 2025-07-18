package injection

import (
	"encoding/json"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/archine/gin-plus/v4/internal/vars/sysconf"
	"github.com/spf13/cast"
)

const (
	// ValueTag is used to mark fields for automatic injection of config values.
	// Note: Unlike the autowire tag, it is only for injecting config values
	ValueTag = "value"
)

var (
	confRegex    = regexp.MustCompile(`\$\{([^:}]+)(?::([^}]*))?}`)
	timeType     = reflect.TypeOf(time.Time{})
	durationType = reflect.TypeOf(time.Duration(0))
)

// CleanInjectCache cleans up the global state of the injection package.
func CleanInjectCache() {
	confRegex = nil
	timeType = nil
	durationType = nil
}

// InjectConfig injects configuration values into struct fields based on the provided tag value.
func InjectConfig(fieldValue reflect.Value, fieldType reflect.Type, tagValue string) {
	submatch := confRegex.FindStringSubmatch(tagValue)
	if len(submatch) == 0 {
		return
	}

	key := submatch[1]
	defaultValue := submatch[2]

	value := sysconf.GlobalProvider.Get(key)
	if value == nil {
		value = defaultValue
	}

	setFieldValue(fieldValue, convert(fieldType, value))
}

// convert converts value to the appropriate type based on the targetType.
func convert(targetType reflect.Type, val any) any {
	if val == nil {
		return nil
	}

	if targetType.Kind() == reflect.Ptr {
		// if targetType is a pointer, we need to dereference it to get the element type
		originTyp := targetType.Elem()

		if elemValue := convert(originTyp, val); elemValue != nil {
			ptr := reflect.New(originTyp)
			ptr.Elem().Set(reflect.ValueOf(elemValue))
			return ptr.Interface()
		}
		return nil
	}

	if targetType == timeType {
		return cast.ToTime(val)
	}
	if targetType == durationType {
		return cast.ToDuration(val)
	}

	switch targetType.Kind() {
	case reflect.Bool:
		return cast.ToBool(val)
	case reflect.String:
		return cast.ToString(val)
	case reflect.Int:
		return cast.ToInt(val)
	case reflect.Int8:
		return cast.ToInt8(val)
	case reflect.Int16:
		return cast.ToInt16(val)
	case reflect.Int32:
		return cast.ToInt32(val)
	case reflect.Int64:
		return cast.ToInt64(val)
	case reflect.Uint:
		return cast.ToUint(val)
	case reflect.Uint8:
		return cast.ToUint16(val)
	case reflect.Uint32:
		return cast.ToUint32(val)
	case reflect.Uint64:
		return cast.ToUint64(val)
	case reflect.Float32:
		return cast.ToFloat32(val)
	case reflect.Float64:
		return cast.ToFloat64(val)
	case reflect.Slice:
		return convertSliceValue(targetType, val)
	default:
		return nil
	}
}

func convertSliceValue(targetType reflect.Type, val any) any {
	if valStr, ok := val.(string); ok {
		valStr = strings.ReplaceAll(strings.TrimSpace(valStr), "'", "\"")
		err := json.Unmarshal([]byte(valStr), &val)
		if err != nil {
			return nil
		}
	}

	elemType := targetType.Elem()
	switch elemType.Kind() {
	case reflect.Bool:
		return cast.ToBoolSlice(val)
	case reflect.String:
		return cast.ToStringSlice(val)
	case reflect.Int:
		return cast.ToIntSlice(val)
	case reflect.Int8:
		v, _ := cast.ToInt8SliceE(val)
		return v
	case reflect.Int16:
		v, _ := cast.ToInt16SliceE(val)
		return v
	case reflect.Int32:
		v, _ := cast.ToInt32SliceE(val)
		return v
	case reflect.Int64:
		return cast.ToInt64Slice(val)
	case reflect.Uint:
		return cast.ToUintSlice(val)
	case reflect.Uint8:
		v, _ := cast.ToUint8SliceE(val)
		return v
	case reflect.Uint16:
		v, _ := cast.ToUint16SliceE(val)
		return v
	case reflect.Uint32:
		v, _ := cast.ToUint32SliceE(val)
		return v
	case reflect.Uint64:
		v, _ := cast.ToUint64SliceE(val)
		return v
	case reflect.Float32:
		v, _ := cast.ToFloat32SliceE(val)
		return v
	case reflect.Float64:
		return cast.ToFloat64Slice(val)
	default:
		return nil
	}
}
