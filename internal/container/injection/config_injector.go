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
	// ValueTag is used to mark fields for automatic injection of gpconf values.
	// Note: Unlike the autowire tag, it is only for injecting gpconf values
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

	value := sysconf.ProjectConfigure.Get(key)
	if value == nil {
		value = defaultValue
	}

	value = convert(fieldType, value)

	setFieldValue(fieldValue, value)
}

// convert converts a string value to the appropriate type based on the targetType.
func convert(targetType reflect.Type, val any) any {
	if targetType.Kind() == reflect.Ptr {
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
	case reflect.Int8, reflect.Int16, reflect.Int32:
		intSlice := cast.ToIntSlice(val)
		return convertIntSlice(intSlice, elemType)
	default:
		return nil
	}
}

func convertIntSlice(intSlice []int, elemType reflect.Type) any {
	switch elemType.Kind() {
	case reflect.Int8:
		result := make([]int8, len(intSlice))
		for i, v := range intSlice {
			result[i] = int8(v)
		}
		return result
	case reflect.Int16:
		result := make([]int16, len(intSlice))
		for i, v := range intSlice {
			result[i] = int16(v)
		}
		return result
	case reflect.Int32:
		result := make([]int32, len(intSlice))
		for i, v := range intSlice {
			result[i] = int32(v)
		}
		return result
	default:
		return intSlice
	}
}
