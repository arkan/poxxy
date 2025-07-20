package poxxy

import (
	"database/sql/driver"
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// RequiredValidator is a special validator that needs access to the schema
type RequiredValidator struct {
	msg string
}

// Validate validates that the field is present and not empty
func (v RequiredValidator) Validate(value interface{}, fieldName string) error {
	// This will be called during the validation phase, but we need schema context
	// The actual logic will be handled in the field's Validate method
	return nil
}

// WithMessage sets a custom error message for the validator
func (v RequiredValidator) WithMessage(msg string) Validator {
	return RequiredValidator{msg: msg}
}

// ValidateWithSchema validates field presence using schema context
func (v RequiredValidator) ValidateWithSchema(schema *Schema, fieldName string) error {
	if !schema.IsFieldPresent(fieldName) {
		if v.msg != "" {
			return fmt.Errorf("%s", v.msg)
		}

		return fmt.Errorf("field is required")
	}

	// Additionally, check that the value is not empty
	value, _ := schema.GetFieldValue(fieldName)
	validator := NotEmpty()
	if err := validator.Validate(value, fieldName); err != nil {
		if v.msg != "" {
			return fmt.Errorf("%s", v.msg)
		}

		return err
	}

	return nil
}

// Required validator - checks if field was present in input data, not if value is non-zero
func Required() Validator {
	return RequiredValidator{}
}

// NotEmpty validator - rejects zero values (use this for non-zero value requirements)
func NotEmpty() Validator {
	return NewInterfaceValidator(func(value interface{}, fieldName string) error {
		if value == nil {
			return fmt.Errorf("field is required")
		}

		// Handle driver.Valuer
		if valuer, ok := value.(driver.Valuer); ok {
			vv, err := valuer.Value()
			if err != nil {
				return fmt.Errorf("error getting value from driver.Valuer: %w", err)
			}

			value = vv
		}

		// Check for zero values
		v := reflect.ValueOf(value)
		switch v.Kind() {
		case reflect.String:
			if v.String() == "" {
				return fmt.Errorf("value cannot be empty")
			}
		case reflect.Slice, reflect.Map:
			if v.Len() == 0 {
				return fmt.Errorf("value cannot be empty")
			}
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			// We cannot refuse zero values for int types.
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			// We cannot refuse zero values for uint types.
		case reflect.Float32, reflect.Float64:
			// We cannot refuse zero values for float types.
		}

		return nil
	})
}

// Email validator validates email format
func Email() Validator {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return NewInterfaceValidator(func(value interface{}, fieldName string) error {
		// Handle driver.Valuer
		if valuer, ok := value.(driver.Valuer); ok {
			vv, err := valuer.Value()
			if err != nil {
				return fmt.Errorf("error getting value from driver.Valuer for: %w", err)
			}
			value = vv
		}
		// If the value is nil, we consider it valid.
		// Use the Required() validator to enforce presence.
		if value == nil {
			return nil
		}

		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("email validation requires string value and not a %T type", value)
		}

		// If the string is empty, we consider it valid.
		// Use the Required() validator to enforce presence.
		if str == "" {
			return nil
		}

		if !emailRegex.MatchString(str) {
			return fmt.Errorf("invalid email format")
		}

		return nil
	})
}

// Min validator validates that a numeric value is at least the specified minimum
func Min(min interface{}) Validator {
	return NewInterfaceValidator(func(value interface{}, fieldName string) error {
		// Handle driver.Valuer
		if valuer, ok := value.(driver.Valuer); ok {
			vv, err := valuer.Value()
			if err != nil {
				return fmt.Errorf("error getting value from driver.Valuer: %w", err)
			}
			value = vv
		}

		v := reflect.ValueOf(value)
		m := reflect.ValueOf(min)

		if m.Kind() != v.Kind() {
			return fmt.Errorf("value must be a %T type", min)
		}

		// Only handle numeric types
		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if v.Int() < m.Convert(v.Type()).Int() {
				return fmt.Errorf("value must be at least %d", m.Int())
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if v.Uint() < m.Convert(v.Type()).Uint() {
				return fmt.Errorf("value must be at least %d", m.Uint())
			}
		case reflect.Float32, reflect.Float64:
			if v.Float() < m.Convert(v.Type()).Float() {
				return fmt.Errorf("value must be at least %f", m.Float())
			}
		default:
			return fmt.Errorf("value must be a numeric type")
		}
		return nil
	})
}

// Max validator validates that a numeric value is at most the specified maximum
func Max(max interface{}) Validator {
	return NewInterfaceValidator(func(value interface{}, fieldName string) error {
		// Handle driver.Valuer
		if valuer, ok := value.(driver.Valuer); ok {
			vv, err := valuer.Value()
			if err != nil {
				return fmt.Errorf("error getting value from driver.Valuer for: %w", err)
			}
			value = vv
		}

		v := reflect.ValueOf(value)
		m := reflect.ValueOf(max)

		if m.Kind() != v.Kind() {
			return fmt.Errorf("value must be a %T type and not a %T type", max, value)
		}

		// Only handle numeric types
		switch v.Kind() {
		case reflect.Int, reflect.Int8, reflect.Int16, reflect.Int32, reflect.Int64:
			if v.Int() > m.Convert(v.Type()).Int() {
				return fmt.Errorf("value must be at most %d", m.Int())
			}
		case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
			if v.Uint() > m.Convert(v.Type()).Uint() {
				return fmt.Errorf("value must be at most %d", m.Uint())
			}
		case reflect.Float32, reflect.Float64:
			if v.Float() > m.Convert(v.Type()).Float() {
				return fmt.Errorf("value must be at most %f", m.Float())
			}
		default:
			return fmt.Errorf("value must be a numeric type")
		}
		return nil
	})
}

// MinLength validator validates that a string or slice has at least the specified length
func MinLength(minLen int) Validator {
	return NewInterfaceValidator(func(value interface{}, fieldName string) error {
		if valuer, ok := value.(driver.Valuer); ok {
			vv, err := valuer.Value()
			if err != nil {
				return fmt.Errorf("error getting value from driver.Valuer for: %w", err)
			}
			value = vv
		}

		v := reflect.ValueOf(value)
		switch v.Kind() {
		case reflect.String:
			if v.Len() < minLen {
				return fmt.Errorf("must be at least %d characters long", minLen)
			}
		case reflect.Slice, reflect.Array:
			if v.Len() < minLen {
				return fmt.Errorf("must have at least %d items", minLen)
			}
		}
		return nil
	})
}

// MaxLength validator validates that a string or slice has at most the specified length
func MaxLength(maxLen int) Validator {
	return NewInterfaceValidator(func(value interface{}, fieldName string) error {
		if valuer, ok := value.(driver.Valuer); ok {
			vv, err := valuer.Value()
			if err != nil {
				return fmt.Errorf("error getting value from driver.Valuer for: %w", err)
			}

			value = vv
		}

		v := reflect.ValueOf(value)
		switch v.Kind() {
		case reflect.String:
			if v.Len() > maxLen {
				return fmt.Errorf("must be at most %d characters long", maxLen)
			}
		case reflect.Slice, reflect.Array:
			if v.Len() > maxLen {
				return fmt.Errorf("must have at most %d items", maxLen)
			}
		}
		return nil
	})
}

// URL validator validates URL format
func URL() Validator {
	return NewInterfaceValidator(func(value interface{}, fieldName string) error {
		if valuer, ok := value.(driver.Valuer); ok {
			vv, err := valuer.Value()
			if err != nil {
				return fmt.Errorf("error getting value from driver.Valuer for: %w", err)
			}

			value = vv
		}

		str, ok := value.(string)
		if !ok {
			return fmt.Errorf("URL validation requires string value")
		}

		// If the string is empty, we consider it valid.
		// Use the Required() validator to enforce presence.
		if str == "" {
			return nil
		}

		if !strings.HasPrefix(str, "http://") && !strings.HasPrefix(str, "https://") {
			return fmt.Errorf("invalid URL format")
		}
		// Check for domain part after protocol
		if str == "http://" || str == "https://" {
			return fmt.Errorf("invalid URL format")
		}

		return nil
	})
}

// ValidatorFunc creates a custom validator from a function (simplified version)
func ValidatorFunc[T any](fn func(value T, fieldName string) error) Validator {
	return NewValidatorFn[T](fn)
}

// In validator validates that a value is one of the specified values
func In(values ...interface{}) Validator {
	return NewInterfaceValidator(func(value interface{}, fieldName string) error {
		for _, v := range values {
			// If value is a driver.Valuer, get the value from it
			if valuer, ok := value.(driver.Valuer); ok {
				vv, err := valuer.Value()
				if err != nil {
					return fmt.Errorf("error getting value from driver.Valuer for: %w", err)
				}

				value = vv
			}

			// We compare the 2 values using reflect.DeepEqual
			if reflect.DeepEqual(value, v) {
				return nil
			}
		}

		return fmt.Errorf("value must be one of: %v", values)
	})
}

// Each validator applies validators to each element of a slice/array
func Each(validators ...Validator) Validator {
	return NewInterfaceValidator(func(value interface{}, fieldName string) error {
		v := reflect.ValueOf(value)
		if v.Kind() != reflect.Slice && v.Kind() != reflect.Array {
			return fmt.Errorf("Each validator can only be applied to slices or arrays")
		}

		for i := 0; i < v.Len(); i++ {
			item := v.Index(i).Interface()
			for _, validator := range validators {
				if err := validator.Validate(item, fmt.Sprintf("%s[%d]", fieldName, i)); err != nil {
					return err
				}
			}
		}

		return nil
	})
}

// Unique validator ensures all elements in slices, arrays, or maps are unique
func Unique() Validator {
	return NewInterfaceValidator(func(value interface{}, fieldName string) error {
		v := reflect.ValueOf(value)

		switch v.Kind() {
		case reflect.Slice, reflect.Array:
			seen := make(map[interface{}]bool)
			for i := 0; i < v.Len(); i++ {
				item := v.Index(i).Interface()
				if seen[item] {
					return fmt.Errorf("duplicate value found: %v", item)
				}
				seen[item] = true
			}

			return nil

		case reflect.Map:
			seen := make(map[interface{}]bool)
			for _, key := range v.MapKeys() {
				mapValue := v.MapIndex(key).Interface()
				if seen[mapValue] {
					return fmt.Errorf("duplicate value found: %v", mapValue)
				}
				seen[mapValue] = true
			}
			return nil

		default:
			return fmt.Errorf("Unique validator can only be applied to slices, arrays, or maps")
		}
	})
}

// UniqueBy validator ensures all elements in slices/arrays are unique by a specific key extractor function
func UniqueBy(keyExtractor func(interface{}) interface{}) Validator {
	return NewInterfaceValidator(func(value interface{}, fieldName string) error {
		v := reflect.ValueOf(value)

		switch v.Kind() {
		case reflect.Slice, reflect.Array:
			seen := make(map[interface{}]bool)
			for i := 0; i < v.Len(); i++ {
				item := v.Index(i).Interface()
				key := keyExtractor(item)
				if seen[key] {
					return fmt.Errorf("duplicate key found: %v", key)
				}
				seen[key] = true
			}
			return nil

		default:
			return fmt.Errorf("UniqueBy validator can only be applied to slices or arrays")
		}
	})
}

// WithMapKeys validator ensures that a map contains all the specified keys
func WithMapKeys(keys ...string) Validator {
	return NewInterfaceValidator(func(value interface{}, fieldName string) error {
		// Try to convert to map[string]string first
		if mapData, ok := value.(map[string]string); ok {
			for _, key := range keys {
				if _, ok := mapData[key]; !ok {
					return fmt.Errorf("key %v not found in map", key)
				}
			}

			return nil
		}
		if mapData, ok := value.(map[string]interface{}); ok {
			for _, key := range keys {
				if _, ok := mapData[key]; !ok {
					return fmt.Errorf("key %v not found in map", key)
				}
			}
		}

		return fmt.Errorf("expected map for map field")
	})
}
