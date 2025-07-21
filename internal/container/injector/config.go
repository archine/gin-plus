package injector

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/archine/gin-plus/v4/internal/container/util"
	"github.com/go-viper/mapstructure/v2"

	"github.com/archine/gin-plus/v4/internal/vars/sysconf"
	"github.com/spf13/cast"
)

const (
	// ValueTag is used to mark fields for automatic injector of config values.
	// Note: Unlike the autowire tag, it is only for injecting config values
	// Example: `value:"${key:default}"`
	ValueTag = "value"
)

var (
	timeType     = reflect.TypeOf(time.Time{})
	durationType = reflect.TypeOf(time.Duration(0))
)

// CleanWireConfigCache clears the cached regular expression and type information used for injecting configuration values.
func CleanWireConfigCache() {
	timeType = nil
	durationType = nil
}

// WireConfigValue injects a configuration value into a struct field based on the provided tag value.
func WireConfigValue(fieldValue reflect.Value, fieldType reflect.Type, tagValue string) error {
	if !strings.HasPrefix(tagValue, "${") || !strings.HasSuffix(tagValue, "}") {
		return fmt.Errorf("invalid config tag value: %s, must be in the format ${key:default}", tagValue)
	}

	content := tagValue[2 : len(tagValue)-1]
	parts := strings.SplitN(content, ":", 2)

	key := parts[0]
	defaultValue := ""
	if len(parts) > 1 {
		defaultValue = strings.TrimSpace(parts[1])
	}

	value := sysconf.Provider.Get(key)
	if value == nil {
		if defaultValue == "" {
			return nil
		}

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

	val = parseJSONIfNeeded(val)

	if convertValue := convertBasic(actualType, val); convertValue != nil {
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

	if err = decoder.Decode(val); err != nil {
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

// parseJSONIfNeeded tries to parse a string as JSON if it looks like a JSON object or array.
func parseJSONIfNeeded(val any) any {
	if valStr, ok := val.(string); ok {
		trimmed := strings.TrimSpace(valStr)
		if (strings.HasPrefix(trimmed, "{") && strings.HasSuffix(trimmed, "}")) ||
			(strings.HasPrefix(trimmed, "[") && strings.HasSuffix(trimmed, "]")) {
			var parsed any
			if err := json.Unmarshal([]byte(valStr), &parsed); err == nil {
				return parsed
			}
		}
	}
	return val
}
