package poxxy

import (
	"fmt"
	"reflect"
	"strconv"
	"strings"
)

// convertValue converts interface{} to type T
func convertValue[T any](value interface{}) (T, error) {
	var zero T

	// Direct type assertion first
	if v, ok := value.(T); ok {
		return v, nil
	}

	// Handle string conversions
	targetType := reflect.TypeOf(zero)
	sourceValue := reflect.ValueOf(value)

	if targetType == sourceValue.Type() {
		return value.(T), nil
	}

	// Convert based on target type
	switch kind := targetType.Kind(); kind {
	case reflect.String:
		str := fmt.Sprintf("%v", value)
		return any(str).(T), nil
	case reflect.Int, reflect.Int64:
		switch v := value.(type) {
		case int:
			if kind == reflect.Int64 {
				return any(int64(v)).(T), nil
			}
			return any(v).(T), nil
		case int64:
			if kind == reflect.Int {
				return any(int(v)).(T), nil
			}
			return any(v).(T), nil
		case float64:
			if kind == reflect.Int64 {
				return any(int64(v)).(T), nil
			}
			return any(int(v)).(T), nil
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				if kind == reflect.Int64 {
					return any(int64(i)).(T), nil
				}

				return any(i).(T), nil
			}
		}
	case reflect.Float64:
		switch v := value.(type) {
		case float64:
			return any(v).(T), nil
		case int:
			return any(float64(v)).(T), nil
		case int64:
			return any(float64(v)).(T), nil
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return any(f).(T), nil
			}
		}
	case reflect.Bool:
		if b, ok := value.(bool); ok {
			return any(b).(T), nil
		}

		if v, ok := value.(string); ok {
			lower := strings.ToLower(v)
			if lower == "true" || lower == "1" || lower == "yes" ||
				lower == "y" || lower == "on" || lower == "t" {
				return any(true).(T), nil
			}

			return any(false).(T), nil
		}
	}

	return zero, fmt.Errorf("cannot convert %T to %T", value, zero)
}
