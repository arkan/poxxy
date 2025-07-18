package poxxy

import (
	"fmt"
	"net/url"
	"regexp"
)

// HTTPMapField represents a map field where each value is a struct
type HTTPMapField[K comparable, V any] struct {
	name        string
	description string
	ptr         *map[K]V
	callback    func(*Schema, *V)
	Validators  []Validator
	wasAssigned bool // Track if a non-nil value was assigned
}

func (f *HTTPMapField[K, V]) Name() string {
	return f.name
}

func (f *HTTPMapField[K, V]) Value() interface{} {
	if f.ptr == nil {
		return nil
	}
	if !f.wasAssigned {
		return nil
	}
	return *f.ptr
}

func (f *HTTPMapField[K, V]) Description() string {
	return f.description
}

func (f *HTTPMapField[K, V]) SetDescription(description string) {
	f.description = description
}

func (f *HTTPMapField[K, V]) Assign(data map[string]interface{}, schema *Schema) error {
	result := make(map[K]V)

	values := convertToURLValues(data)
	formData := parseFormCollection(values, f.name)
	if len(formData) > 0 {
		schema.SetFieldPresent(f.name)
		f.wasAssigned = true
	} else {
		f.wasAssigned = false
	}

	for key, value := range formData {
		convertedKey, err := convertValue[K](key)
		if err != nil {
			return fmt.Errorf("key %s: failed to convert: %v", key, err)
		}

		var element V
		subSchema := NewSchema()
		f.callback(subSchema, &element)
		if err := subSchema.Apply(convertMapStringStringToMapStringInterface(value)); err != nil {
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

// AppendValidators implements ValidatorsAppender interface
func (f *HTTPMapField[K, V]) AppendValidators(validators []Validator) {
	f.Validators = append(f.Validators, validators...)
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

func parseFormCollection(values url.Values, typeName string) map[string]map[string]string {
	result := make(map[string]map[string]string)

	for key, values := range values {
		re := regexp.MustCompile(typeName + "\\[([0-9A-Za-z_-]+)\\]\\[([a-zA-Z_-]+)\\]")
		matches := re.FindStringSubmatch(key)

		if len(matches) >= 3 {
			identifier := matches[1]
			fieldName := matches[2]

			// Initialize the inner map if it doesn't exist
			if _, exists := result[identifier]; !exists {
				result[identifier] = make(map[string]string)
			}

			result[identifier][fieldName] = values[0]
		}
	}
	return result
}

func convertMapStringStringToMapStringInterface(data map[string]string) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range data {
		result[key] = value
	}
	return result
}
