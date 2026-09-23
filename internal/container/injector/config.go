package injector

import (
	"encoding/json"
	"fmt"
	"reflect"
	"strings"
	"time"

	"github.com/archine/gin-plus/v4/component/config"
	"github.com/archine/gin-plus/v4/internal/container/util"
	"github.com/go-viper/mapstructure/v2"
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

// ValueExpr is the parsed form of a `value:"${key:default}"` tag.
type ValueExpr struct {
	Key      string // Config key to look up
	Default  string // Fallback literal used when the key is absent; empty means "keep the zero value"
	Required bool   // Set by the "?" sentinel (`${key:?}`): an absent key fails start-up
}

// ParseValueExpr parses a value tag of the form ${key}, ${key:default} or ${key:?}.
func ParseValueExpr(tag string) (ValueExpr, error) {
	if !strings.HasPrefix(tag, "${") || !strings.HasSuffix(tag, "}") {
		return ValueExpr{}, fmt.Errorf("invalid value tag '%s', expected format ${key:default}", tag)
	}

	key, def, _ := strings.Cut(tag[2:len(tag)-1], ":")
	expr := ValueExpr{Key: strings.TrimSpace(key), Default: strings.TrimSpace(def)}
	if expr.Key == "" {
		return ValueExpr{}, fmt.Errorf("invalid value tag '%s', config key is empty", tag)
	}
	if expr.Default == "?" {
		expr.Default, expr.Required = "", true
	}
	return expr, nil
}

// WireConfigValue looks up expr in cp and assigns the converted value to fieldValue.
// fieldValue must be addressable; unexported fields are supported.
func WireConfigValue(cp config.Provider, fieldValue reflect.Value, expr ValueExpr) error {
	val := cp.Get(expr.Key)
	if val == nil {
		switch {
		case expr.Required:
			return fmt.Errorf("required config key '%s' is missing", expr.Key)
		case expr.Default == "":
			return nil
		}
		val = expr.Default
	}

	if err := setField(fieldValue, val); err != nil {
		return fmt.Errorf("config key '%s': %w", expr.Key, err)
	}
	return nil
}

// setField is the single dispatch point: it resolves the concrete target type and
// routes to the appropriate converter. The target type drives all decisions — no
// speculative parsing is done before we know what we are converting into.
func setField(fieldValue reflect.Value, val any) error {
	fieldType := fieldValue.Type()
	actualType := fieldType
	isPtr := fieldType.Kind() == reflect.Pointer
	if isPtr {
		actualType = fieldType.Elem()
	}

	result, err := convert(actualType, val)
	if err != nil {
		return err
	}

	if isPtr {
		ptr := reflect.New(actualType)
		ptr.Elem().Set(result)
		result = ptr
	}
	util.DirectSetValue(fieldValue, result)
	return nil
}

// convert produces a value of exactly targetType (never a pointer).
// Basic kinds go through cast; the result is converted to named types such as
// `type Env string`. Anything cast cannot produce directly (structs, maps, slices
// of structs or of named elements) is decoded by mapstructure.
func convert(targetType reflect.Type, val any) (reflect.Value, error) {
	converted, err := dispatchConvert(targetType, val)
	if err != nil {
		return reflect.Value{}, err
	}

	if converted != nil {
		rv := reflect.ValueOf(converted)
		if rv.Type() == targetType {
			return rv, nil
		}
		if rv.Type().ConvertibleTo(targetType) {
			return rv.Convert(targetType), nil
		}
		// e.g. []string → []Env is not convertible; let mapstructure decode element-wise.
	}

	return decodeComplex(targetType, val)
}

// dispatchConvert converts val to the basic/primitive kind of targetType.
// Returns (nil, nil) to signal "not a basic type — caller should use decodeComplex".
// The returned value has the unnamed type of the kind (e.g. string for `type Env string`).
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
