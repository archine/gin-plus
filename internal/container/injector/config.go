package injector

import (
	"encoding/json"
	"github.com/archine/gin-plus/v4/internal/container/util"
	"github.com/go-viper/mapstructure/v2"
	"reflect"
	"regexp"
	"strings"
	"time"

	"github.com/archine/gin-plus/v4/internal/vars/sysconf"
	"github.com/spf13/cast"
)

const (
	// ValueTag is used to mark fields for automatic injector of config values.
	// Note: Unlike the autowire tag, it is only for injecting config values
	ValueTag = "value"
)

var (
	confRegex    = regexp.MustCompile(`\$\{([^:}]+)(?::([^}]*))?}`)
	timeType     = reflect.TypeOf(time.Time{})
	durationType = reflect.TypeOf(time.Duration(0))
)

// CleanWireConfigCache clears the cached regular expression and type information used for injecting configuration values.
func CleanWireConfigCache() {
	confRegex = nil
	timeType = nil
	durationType = nil
}

// WireConfigValue injects a configuration value into a struct field based on the provided tag value.
func WireConfigValue(fieldValue reflect.Value, fieldType reflect.Type, tagValue string) error {
	submatch := confRegex.FindStringSubmatch(tagValue)
	if len(submatch) == 0 {
		return nil
	}

	key := submatch[1]
	defaultValue := submatch[2]

	value := sysconf.Provider.Get(key)
	if value == nil {
		if defaultValue == "" {
			return nil // No value found and no default provided, nothing to do
		}
		defaultValue = strings.ReplaceAll(strings.TrimSpace(defaultValue), "'", "\"")
		value = defaultValue
	}

	return convert(fieldValue, fieldType, value)
}

// convert converts value to the appropriate type and sets it to fieldValue
func convert(fieldValue reflect.Value, fieldType reflect.Type, val any) error {
	actualType := fieldType
	if fieldType.Kind() == reflect.Ptr {
		actualType = fieldType.Elem()
	}
	if convertValue := convertBasic(actualType, val); convertValue != nil {
		// If the field is a pointer, we need to set the value to the pointer
		if fieldValue.Kind() == reflect.Ptr {
			newPtr := reflect.New(actualType)
			newPtr.Elem().Set(reflect.ValueOf(convertValue))
			util.DirectSetValue(fieldValue, newPtr)
		} else {
			util.DirectSetValue(fieldValue, reflect.ValueOf(convertValue))
		}
		return nil
	}

	return convertOther(fieldValue, val)
}

// convertBasic handles basic types
func convertBasic(targetType reflect.Type, val any) any {
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
		return cast.ToUint8(val)
	case reflect.Uint16:
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

// convertOther handles complex types using map-structure
func convertOther(fieldValue reflect.Value, val any) error {
	tempValue := fieldValue
	if !fieldValue.CanSet() {
		tempValue = reflect.New(fieldValue.Type())
	}
	if varStr, ok := val.(string); ok {
		// If the value is a string, we need to parse it
		_ = json.Unmarshal([]byte(varStr), &val)
	}

	config := &mapstructure.DecoderConfig{
		Result:           tempValue.Interface(),
		WeaklyTypedInput: true,
		DecodeHook: mapstructure.ComposeDecodeHookFunc(
			mapstructure.StringToTimeDurationHookFunc(),
			mapstructure.StringToSliceHookFunc(","),
		),
	}

	decoder, err := mapstructure.NewDecoder(config)
	if err != nil {
		return err
	}

	err = decoder.Decode(val)
	if err != nil {
		return err
	}

	util.DirectSetValue(fieldValue, tempValue.Elem())
	return nil
}

// convertSliceValue converts a value to a slice of the specified type.
func convertSliceValue(targetType reflect.Type, val any) any {
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
