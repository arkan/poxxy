package poxxy

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"
)

// RequiredValidator is a special validator that needs access to the schema
type RequiredValidator struct {
	msg string
}

func (v RequiredValidator) Validate(value interface{}, fieldName string) error {
	// This will be called during the validation phase, but we need schema context
	// The actual logic will be handled in the field's Validate method
	return nil
}

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
	return nil
}

// Required validator - checks if field was present in input data, not if value is non-zero
func Required() Validator {
	return RequiredValidator{}
}

// NonZero validator - rejects zero values (use this for non-zero value requirements)
func NonZero() Validator {
	return ValidatorFn{
		fn: func(value interface{}, fieldName string) error {
			if value == nil {
				return fmt.Errorf("value cannot be zero")
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
				if v.Int() == 0 {
					return fmt.Errorf("value cannot be zero")
				}
			case reflect.Uint, reflect.Uint8, reflect.Uint16, reflect.Uint32, reflect.Uint64:
				if v.Uint() == 0 {
					return fmt.Errorf("value cannot be zero")
				}
			case reflect.Float32, reflect.Float64:
				if v.Float() == 0.0 {
					return fmt.Errorf("value cannot be zero")
				}
			}

			return nil
		},
	}
}

// Email validator
func Email() Validator {
	emailRegex := regexp.MustCompile(`^[a-zA-Z0-9._%+-]+@[a-zA-Z0-9.-]+\.[a-zA-Z]{2,}$`)
	return ValidatorFn{
		fn: func(value interface{}, fieldName string) error {
			str, ok := value.(string)
			if !ok {
				return fmt.Errorf("email validation requires string value")
			}
			if !emailRegex.MatchString(str) {
				return fmt.Errorf("invalid email format")
			}
			return nil
		},
	}
}

// Min validator
func Min(min interface{}) Validator {
	return ValidatorFn{
		fn: func(value interface{}, fieldName string) error {
			switch v := value.(type) {
			case int:
				if minInt, ok := min.(int); ok && v < minInt {
					return fmt.Errorf("value must be at least %d", minInt)
				}
			case float64:
				if minFloat, ok := min.(float64); ok && v < minFloat {
					return fmt.Errorf("value must be at least %f", minFloat)
				}
			}
			return nil
		},
	}
}

// Max validator
func Max(max interface{}) Validator {
	return ValidatorFn{
		fn: func(value interface{}, fieldName string) error {
			switch v := value.(type) {
			case int:
				if maxInt, ok := max.(int); ok && v > maxInt {
					return fmt.Errorf("value must be at most %d", maxInt)
				}
			case float64:
				if maxFloat, ok := max.(float64); ok && v > maxFloat {
					return fmt.Errorf("value must be at most %f", maxFloat)
				}
			}
			return nil
		},
	}
}

// MinLength validator
func MinLength(minLen int) Validator {
	return ValidatorFn{
		fn: func(value interface{}, fieldName string) error {
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
		},
	}
}

// MaxLength validator
func MaxLength(maxLen int) Validator {
	return ValidatorFn{
		fn: func(value interface{}, fieldName string) error {
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
		},
	}
}

// URL validator
func URL() Validator {
	return ValidatorFn{
		fn: func(value interface{}, fieldName string) error {
			str, ok := value.(string)
			if !ok {
				return fmt.Errorf("URL validation requires string value")
			}
			if !strings.HasPrefix(str, "http://") && !strings.HasPrefix(str, "https://") {
				return fmt.Errorf("invalid URL format")
			}
			// Check for domain part after protocol
			if str == "http://" || str == "https://" {
				return fmt.Errorf("invalid URL format")
			}
			return nil
		},
	}
}

func ValidatorFunc[T any](fn func(value T, fieldName string) error) Validator {
	return ValidatorFn{
		fn: func(value interface{}, fieldName string) error {
			return fn(value.(T), fieldName)
		},
	}
}

// In validator
func In(values ...interface{}) Validator {
	return ValidatorFn{
		fn: func(value interface{}, fieldName string) error {
			for _, v := range values {
				if reflect.DeepEqual(value, v) {
					return nil
				}
			}
			return fmt.Errorf("value must be one of: %v", values)
		},
	}
}

// Each validator applies validators to each element of a slice/array
func Each(validators ...Validator) Validator {
	return ValidatorFn{
		fn: func(value interface{}, fieldName string) error {
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
		},
	}
}

// Unique validator ensures all elements in slices, arrays, or maps are unique
func Unique() Validator {
	return ValidatorFn{
		fn: func(value interface{}, fieldName string) error {
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
		},
	}
}

// UniqueBy validator ensures all elements in slices/arrays are unique by a specific key extractor function
func UniqueBy(keyExtractor func(interface{}) interface{}) Validator {
	return ValidatorFn{
		fn: func(value interface{}, fieldName string) error {
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
		},
	}
}

func WithMapKeys(keys ...string) Validator {
	return ValidatorFn{
		fn: func(value interface{}, fieldName string) error {
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
		},
	}
}
