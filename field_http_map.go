package poxxy

import (
	"fmt"
	"net/url"
	"regexp"
)

// HTTPMapField represents a map field where each value is a struct
type HTTPMapField[K comparable, V any] struct {
	name       string
	ptr        *map[K]V
	callback   func(*Schema, *V)
	Validators []Validator
}

func (f *HTTPMapField[K, V]) Name() string {
	return f.name
}

func (f *HTTPMapField[K, V]) Assign(data map[string]interface{}, schema *Schema) error {
	result := make(map[K]V)

	values := convertToURLValues(data)
	formData := parseFormCollection(values, f.name)
	if len(formData) > 0 {
		schema.SetFieldPresent(f.name)
	}

	for key, value := range formData {
		convertedKey, err := convertValue[K](key)
		if err != nil {
			return fmt.Errorf("key %s: failed to convert: %v", key, err)
		}

		var element V
		subSchema := NewSchema()
		f.callback(subSchema, &element)
		if err := subSchema.Apply(value); err != nil {
			return fmt.Errorf("key %s: %v", key, err)
		}
		result[convertedKey] = element
	}

	*f.ptr = result
	return nil
}

func (f *HTTPMapField[K, V]) Validate(schema *Schema) error {
	return validateFieldValidators(f.Validators, *f.ptr, f.name, schema)
}

// HTTPMap creates a map field for structs with element-wise schema definition
func HTTPMap[K comparable, V any](name string, ptr *map[K]V, schemaCallback func(*Schema, *V), opts ...Option) Field {
	field := &HTTPMapField[K, V]{
		name:     name,
		ptr:      ptr,
		callback: schemaCallback,
	}

	for _, opt := range opts {
		opt.Apply(field)
	}

	return field
}

func convertToURLValues(data map[string]interface{}) url.Values {
	values := url.Values{}
	for key, value := range data {
		switch v := value.(type) {
		case []string:
			if len(v) > 0 {
				values.Add(key, v[0])
			} else {
				values.Add(key, "")
			}
		case string:
			values.Add(key, v)
		default:
			values.Add(key, fmt.Sprintf("%v", v))
		}
	}
	return values
}

func parseFormCollection(values url.Values, typeName string) map[string]map[string]interface{} {
	result := make(map[string]map[string]interface{})

	for key, values := range values {
		re := regexp.MustCompile(typeName + "\\[([0-9A-Za-z_-]+)\\]\\[([a-zA-Z_-]+)\\]")
		matches := re.FindStringSubmatch(key)

		if len(matches) >= 3 {
			identifier := matches[1]
			fieldName := matches[2]

			// Initialize the inner map if it doesn't exist
			if _, exists := result[identifier]; !exists {
				result[identifier] = make(map[string]interface{})
			}

			result[identifier][fieldName] = values[0]
		}
	}
	return result
}
