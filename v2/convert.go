package poxxy

import (
	"fmt"
	"reflect"
	"strconv"
)

// convertValue converts a raw value to type T.
// If strict is true, only exact type matches are allowed.
func convertValue[T any](raw any, strict bool) (T, error) {
	var zero T

	if raw == nil {
		return zero, nil
	}

	// Direct type assertion
	if v, ok := raw.(T); ok {
		return v, nil
	}

	if strict {
		return zero, fmt.Errorf("expected %T, got %T", zero, raw)
	}

	// Flexible conversion
	return flexibleConvert[T](raw)
}

// flexibleConvert performs automatic type conversion.
func flexibleConvert[T any](raw any) (T, error) {
	var zero T
	targetType := reflect.TypeOf(zero)

	// Handle pointer types
	if targetType == nil {
		// T is interface{}, return as-is
		if v, ok := raw.(T); ok {
			return v, nil
		}
		return zero, fmt.Errorf("cannot convert %T to target type", raw)
	}

	// Get the kind we're converting to
	kind := targetType.Kind()

	// Handle conversions based on target type
	switch kind {
	case reflect.String:
		return convertToString[T](raw)
	case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
		return convertToInt[T](raw, kind)
	case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
		return convertToUint[T](raw, kind)
	case reflect.Float32, reflect.Float64:
		return convertToFloat[T](raw, kind)
	case reflect.Bool:
		return convertToBool[T](raw)
	default:
		return zero, fmt.Errorf("unsupported conversion from %T to %s", raw, targetType)
	}
}

func convertToString[T any](raw any) (T, error) {
	var zero T
	var result string

	switch v := raw.(type) {
	case string:
		result = v
	case []byte:
		result = string(v)
	case int:
		result = strconv.FormatInt(int64(v), 10)
	case int64:
		result = strconv.FormatInt(v, 10)
	case float64:
		result = strconv.FormatFloat(v, 'f', -1, 64)
	case bool:
		result = strconv.FormatBool(v)
	default:
		result = fmt.Sprintf("%v", v)
	}

	// Convert result to T
	rv := reflect.ValueOf(&zero).Elem()
	rv.SetString(result)
	return zero, nil
}

func convertToInt[T any](raw any, kind reflect.Kind) (T, error) {
	var zero T
	var intVal int64

	switch v := raw.(type) {
	case int:
		intVal = int64(v)
	case int8:
		intVal = int64(v)
	case int16:
		intVal = int64(v)
	case int32:
		intVal = int64(v)
	case int64:
		intVal = v
	case uint:
		intVal = int64(v)
	case uint8:
		intVal = int64(v)
	case uint16:
		intVal = int64(v)
	case uint32:
		intVal = int64(v)
	case uint64:
		intVal = int64(v)
	case float32:
		intVal = int64(v)
	case float64:
		intVal = int64(v)
	case string:
		parsed, err := strconv.ParseInt(v, 10, 64)
		if err != nil {
			// Try parsing as float then convert
			f, ferr := strconv.ParseFloat(v, 64)
			if ferr != nil {
				return zero, fmt.Errorf("cannot convert string %q to integer", v)
			}
			intVal = int64(f)
		} else {
			intVal = parsed
		}
	case bool:
		if v {
			intVal = 1
		} else {
			intVal = 0
		}
	default:
		return zero, fmt.Errorf("cannot convert %T to integer", raw)
	}

	rv := reflect.ValueOf(&zero).Elem()
	if rv.OverflowInt(intVal) {
		return zero, fmt.Errorf("integer overflow: %d", intVal)
	}
	rv.SetInt(intVal)
	return zero, nil
}

func convertToUint[T any](raw any, kind reflect.Kind) (T, error) {
	var zero T
	var uintVal uint64

	switch v := raw.(type) {
	case uint:
		uintVal = uint64(v)
	case uint8:
		uintVal = uint64(v)
	case uint16:
		uintVal = uint64(v)
	case uint32:
		uintVal = uint64(v)
	case uint64:
		uintVal = v
	case int:
		if v < 0 {
			return zero, fmt.Errorf("cannot convert negative value %d to unsigned", v)
		}
		uintVal = uint64(v)
	case int64:
		if v < 0 {
			return zero, fmt.Errorf("cannot convert negative value %d to unsigned", v)
		}
		uintVal = uint64(v)
	case float32:
		if v < 0 {
			return zero, fmt.Errorf("cannot convert negative value %f to unsigned", v)
		}
		uintVal = uint64(v)
	case float64:
		if v < 0 {
			return zero, fmt.Errorf("cannot convert negative value %f to unsigned", v)
		}
		uintVal = uint64(v)
	case string:
		parsed, err := strconv.ParseUint(v, 10, 64)
		if err != nil {
			return zero, fmt.Errorf("cannot convert string %q to unsigned integer", v)
		}
		uintVal = parsed
	case bool:
		if v {
			uintVal = 1
		} else {
			uintVal = 0
		}
	default:
		return zero, fmt.Errorf("cannot convert %T to unsigned integer", raw)
	}

	rv := reflect.ValueOf(&zero).Elem()
	if rv.OverflowUint(uintVal) {
		return zero, fmt.Errorf("unsigned integer overflow: %d", uintVal)
	}
	rv.SetUint(uintVal)
	return zero, nil
}

func convertToFloat[T any](raw any, kind reflect.Kind) (T, error) {
	var zero T
	var floatVal float64

	switch v := raw.(type) {
	case float32:
		floatVal = float64(v)
	case float64:
		floatVal = v
	case int:
		floatVal = float64(v)
	case int64:
		floatVal = float64(v)
	case uint:
		floatVal = float64(v)
	case uint64:
		floatVal = float64(v)
	case string:
		parsed, err := strconv.ParseFloat(v, 64)
		if err != nil {
			return zero, fmt.Errorf("cannot convert string %q to float", v)
		}
		floatVal = parsed
	default:
		return zero, fmt.Errorf("cannot convert %T to float", raw)
	}

	rv := reflect.ValueOf(&zero).Elem()
	if rv.OverflowFloat(floatVal) {
		return zero, fmt.Errorf("float overflow: %f", floatVal)
	}
	rv.SetFloat(floatVal)
	return zero, nil
}

func convertToBool[T any](raw any) (T, error) {
	var zero T
	var boolVal bool

	switch v := raw.(type) {
	case bool:
		boolVal = v
	case int:
		boolVal = v != 0
	case int64:
		boolVal = v != 0
	case float64:
		boolVal = v != 0
	case string:
		switch v {
		case "true", "1", "yes", "on":
			boolVal = true
		case "false", "0", "no", "off", "":
			boolVal = false
		default:
			return zero, fmt.Errorf("cannot convert string %q to bool", v)
		}
	default:
		return zero, fmt.Errorf("cannot convert %T to bool", raw)
	}

	rv := reflect.ValueOf(&zero).Elem()
	rv.SetBool(boolVal)
	return zero, nil
}
