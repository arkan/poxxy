package poxxy

import (
	"fmt"
	"net/url"
	"regexp"
)

// HTTPMapField represents a map field where each value is a struct
type HTTPMapField[K comparable, V any] struct {
	name         string
	description  string
	ptr          *map[K]V
	callback     func(*Schema, *V)
	Validators   []Validator
	wasAssigned  bool // Track if a non-nil value was assigned
	defaultValue map[K]V
	hasDefault   bool
}

// Name returns the field name
func (f *HTTPMapField[K, V]) Name() string {
	return f.name
}

// Value returns the current value of the field
func (f *HTTPMapField[K, V]) Value() interface{} {
	if f.ptr == nil {
		return nil
	}

	if !f.wasAssigned {
		return nil
	}

	return *f.ptr
}

// Description returns the field description
func (f *HTTPMapField[K, V]) Description() string {
	return f.description
}

// SetDescription sets the field description
func (f *HTTPMapField[K, V]) SetDescription(description string) {
	f.description = description
}

// Assign assigns a value to the field from the input data
func (f *HTTPMapField[K, V]) Assign(data map[string]interface{}, schema *Schema) error {
	result := make(map[K]V)

	values := convertToURLValues(data)
	formData := parseFormCollection(values, f.name)
	if len(formData) > 0 {
		schema.SetFieldPresent(f.name)
		f.wasAssigned = true
	} else {
		// Apply default value if available and no form data was found
		if f.hasDefault {
			*f.ptr = f.defaultValue
			f.wasAssigned = true
			schema.SetFieldPresent(f.name)
		} else {
			f.wasAssigned = false
		}

		return nil
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

// Validate validates the field value using all registered validators
func (f *HTTPMapField[K, V]) Validate(schema *Schema) error {
	return validateFieldValidators(f.Validators, *f.ptr, f.name, schema)
}

// AppendValidators implements ValidatorsAppender interface
func (f *HTTPMapField[K, V]) AppendValidators(validators []Validator) {
	f.Validators = append(f.Validators, validators...)
}

// SetCallback sets the callback function for configuring sub-schemas
func (f *HTTPMapField[K, V]) SetCallback(callback func(*Schema, *V)) {
	f.callback = callback
}

// SetDefaultValue sets the default value for the field
func (f *HTTPMapField[K, V]) SetDefaultValue(defaultValue map[K]V) {
	f.defaultValue = defaultValue
	f.hasDefault = true
}

// HTTPMapCallbackOption holds a callback function for HTTPMap
type HTTPMapCallbackOption[K comparable, V any] struct {
	callback func(*Schema, *V)
}

// Apply applies the callback to the HTTPMap field
func (o HTTPMapCallbackOption[K, V]) Apply(field interface{}) {
	if httpMapField, ok := field.(*HTTPMapField[K, V]); ok {
		httpMapField.SetCallback(o.callback)
	}
}

// WithHTTPMapCallback creates a callback option for HTTPMap
func WithHTTPMapCallback[K comparable, V any](callback func(*Schema, *V)) Option {
	return HTTPMapCallbackOption[K, V]{callback: callback}
}

// HTTPMap creates an HTTP map field
func HTTPMap[K comparable, V any](name string, ptr *map[K]V, opts ...Option) Field {
	field := &HTTPMapField[K, V]{
		name: name,
		ptr:  ptr,
	}

	for _, opt := range opts {
		opt.Apply(field)
	}

	return field
}

// convertToURLValues converts a map[string]interface{} to url.Values
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

// parseFormCollection parses form data to extract nested map structures
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

// convertMapStringStringToMapStringInterface converts map[string]string to map[string]interface{}
func convertMapStringStringToMapStringInterface(data map[string]string) map[string]interface{} {
	result := make(map[string]interface{})
	for key, value := range data {
		result[key] = value
	}

	return result
}
