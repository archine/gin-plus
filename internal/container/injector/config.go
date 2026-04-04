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
	// ValueTag is used to mark fields for automatic injection of config values.
	// Note: Unlike the autowire tag, it is only for injecting config values.
	// Example: `value:"${key:default}"`
	ValueTag = "value"

	// DecodeTag is used for mapstructure decoding of complex types (structs, maps, slices).
	// It allows both "mapstructure" and "json" tags to be used for field names.
	DecodeTag = "json,mapstructure"
)

var (
	timeType     = reflect.TypeFor[time.Time]()
	durationType = reflect.TypeFor[time.Duration]()

	// complexDecodeHook is stateless and reused across all decodeComplex calls.
	complexDecodeHook = mapstructure.ComposeDecodeHookFunc(
		mapstructure.TextUnmarshallerHookFunc(),
		mapstructure.StringToTimeDurationHookFunc(),
		mapstructure.StringToTimeLocationHookFunc(),
		mapstructure.StringToSliceHookFunc(","),
	)
)

// WireConfigValue resolves the tag expression, looks up the config value, and sets the field.
func WireConfigValue(fieldValue reflect.Value, fieldType reflect.Type, tagValue string) error {
	if !strings.HasPrefix(tagValue, "${") || !strings.HasSuffix(tagValue, "}") {
		return fmt.Errorf("invalid config tag value: %s, must be in the format ${key:default}", tagValue)
	}

	content := tagValue[2 : len(tagValue)-1]
	parts := strings.SplitN(content, ":", 2)

	key := strings.TrimSpace(parts[0])
	defaultVal := ""
	if len(parts) > 1 {
		defaultVal = strings.TrimSpace(parts[1])
	}

	val := sysconf.Provider.Get(key)
	if val == nil {
		// "?" is the required-value sentinel: `value:"${key:?}"` means the key must
		// be present in config; start-up fails if it is absent.
		if defaultVal == "?" {
			return fmt.Errorf("required config key '%s' is missing", key)
		}
		if defaultVal == "" {
			return nil
		}
		val = defaultVal
	}

	return setField(fieldValue, fieldType, val)
}

// setField is the single dispatch point: it resolves the concrete target type and
// routes to the appropriate converter. The target type drives all decisions — no
// speculative parsing is done before we know what we are converting into.
func setField(fieldValue reflect.Value, fieldType reflect.Type, prepareSetValue any) error {
	actualType := fieldType
	isPtr := fieldType.Kind() == reflect.Pointer
	if isPtr {
		actualType = fieldType.Elem()
	}

	converted, err := dispatchConvert(actualType, prepareSetValue)
	if err != nil {
		return err
	}

	if converted != nil {
		if isPtr {
			ptr := reflect.New(actualType)
			ptr.Elem().Set(reflect.ValueOf(converted))
			util.DirectSetValue(fieldValue, ptr)
		} else {
			util.DirectSetValue(fieldValue, reflect.ValueOf(converted))
		}
		return nil
	}

	// Fall through to mapstructure for struct / map / slice-of-struct, etc.
	// JSON pre-parsing happens here, where we actually need it.
	// Always pass the dereferenced type so decodeComplex never sees a pointer.
	result, err := decodeComplex(actualType, prepareSetValue)
	if err != nil {
		return err
	}
	if isPtr {
		util.DirectSetValue(fieldValue, result.Addr())
	} else {
		util.DirectSetValue(fieldValue, result)
	}

	return nil
}

// dispatchConvert converts val to the basic/primitive target type.
// Returns (nil, nil) to signal "not a basic type — caller should use decodeComplex".
func dispatchConvert(targetType reflect.Type, val any) (any, error) {
	// Named types based on reflect.Type identity
	if targetType == timeType {
		return cast.ToTimeInDefaultLocationE(val, nil)
	}
	if targetType == durationType {
		return cast.ToDurationE(val)
	}

	switch targetType.Kind() {
	case reflect.String:
		return cast.ToStringE(val)
	case reflect.Bool:
		return cast.ToBoolE(val)
	case reflect.Int:
		return cast.ToIntE(val)
	case reflect.Int8:
		return cast.ToInt8E(val)
	case reflect.Int16:
		return cast.ToInt16E(val)
	case reflect.Int32:
		return cast.ToInt32E(val)
	case reflect.Int64:
		return cast.ToInt64E(val)
	case reflect.Uint:
		return cast.ToUintE(val)
	case reflect.Uint8:
		return cast.ToUint8E(val)
	case reflect.Uint16:
		return cast.ToUint16E(val)
	case reflect.Uint32:
		return cast.ToUint32E(val)
	case reflect.Uint64:
		return cast.ToUint64E(val)
	case reflect.Float32:
		return cast.ToFloat32E(val)
	case reflect.Float64:
		return cast.ToFloat64E(val)
	case reflect.Slice:
		return dispatchSliceConvert(targetType, val)
	default:
		return nil, nil
	}
}

// dispatchSliceConvert handles slices whose element type is a basic type.
// Returns (nil, nil) for slices of structs / maps, which fall through to decodeComplex.
func dispatchSliceConvert(targetType reflect.Type, val any) (any, error) {
	switch targetType.Elem().Kind() {
	case reflect.String:
		return cast.ToStringSliceE(val)
	case reflect.Bool:
		return cast.ToBoolSliceE(val)
	case reflect.Int:
		return cast.ToIntSliceE(val)
	case reflect.Int8:
		return cast.ToInt8SliceE(val)
	case reflect.Int16:
		return cast.ToInt16SliceE(val)
	case reflect.Int32:
		return cast.ToInt32SliceE(val)
	case reflect.Int64:
		return cast.ToInt64SliceE(val)
	case reflect.Uint:
		return cast.ToUintSliceE(val)
	case reflect.Uint8:
		return cast.ToUint8SliceE(val)
	case reflect.Uint16:
		return cast.ToUint16SliceE(val)
	case reflect.Uint32:
		return cast.ToUint32SliceE(val)
	case reflect.Uint64:
		return cast.ToUint64SliceE(val)
	case reflect.Float32:
		return cast.ToFloat32SliceE(val)
	case reflect.Float64:
		return cast.ToFloat64SliceE(val)
	default:
		return nil, nil
	}
}

// decodeComplex handles struct, map, and other composite types via mapstructure.
// If val is a JSON string (object or array), it is parsed first so that mapstructure
// receives a map/slice rather than a raw string.
// targetType must be the dereferenced (non-pointer) type.
func decodeComplex(targetType reflect.Type, val any) (reflect.Value, error) {
	decoded, err := tryParseJSONString(val)
	if err != nil {
		return reflect.Value{}, err
	}

	ptr := reflect.New(targetType)
	cfg := &mapstructure.DecoderConfig{
		Result:           ptr.Interface(),
		TagName:          DecodeTag,
		WeaklyTypedInput: true,
		DecodeHook:       complexDecodeHook,
	}

	dec, err := mapstructure.NewDecoder(cfg)
	if err != nil {
		return reflect.Value{}, err
	}
	if err = dec.Decode(decoded); err != nil {
		return reflect.Value{}, err
	}

	return ptr.Elem(), nil
}

// tryParseJSONString unmarshal val if it is a string starting with '{' or '['.
// Plain strings are returned unchanged, and non-string values pass through as-is.
// An error is returned only when the input looks like JSON but is malformed.
func tryParseJSONString(val any) (any, error) {
	s, ok := val.(string)
	if !ok {
		return val, nil
	}
	s = strings.TrimSpace(s)
	if len(s) == 0 || (s[0] != '{' && s[0] != '[') {
		return val, nil
	}
	var out any
	if err := json.Unmarshal([]byte(s), &out); err != nil {
		return nil, fmt.Errorf("failed to parse config value as JSON: %w", err)
	}
	return out, nil
}
