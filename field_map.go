package poxxy

import (
	"fmt"
)

// MapField represents a map field
type MapField[K comparable, V any] struct {
	name         string
	description  string
	ptr          *map[K]V
	callback     func(*Schema, K, V)
	Validators   []Validator
	wasAssigned  bool // Track if a non-nil value was assigned
	defaultValue map[K]V
	hasDefault   bool
}

// Name returns the field name
func (f *MapField[K, V]) Name() string {
	return f.name
}

// Value returns the current value of the field
func (f *MapField[K, V]) Value() interface{} {
	if f.ptr == nil {
		return nil
	}

	if !f.wasAssigned {
		return nil
	}

	return *f.ptr
}

// Description returns the field description
func (f *MapField[K, V]) Description() string {
	return f.description
}

// SetDescription sets the field description
func (f *MapField[K, V]) SetDescription(description string) {
	f.description = description
}

// Assign assigns a value to the field from the input data
func (f *MapField[K, V]) Assign(data map[string]interface{}, schema *Schema) error {
	value, exists := data[f.name]
	if !exists {
		// Apply default value if available
		if f.hasDefault {
			*f.ptr = f.defaultValue
			f.wasAssigned = true
			schema.SetFieldPresent(f.name)
		}
		return nil
	}

	schema.SetFieldPresent(f.name)

	if value == nil {
		f.wasAssigned = false
		return nil
	}

	if _, ok := value.(string); ok && value.(string) == "" {
		f.wasAssigned = false
		return nil
	}

	mapData, ok := value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected map for map field")
	}

	result := make(map[K]V)

	for key, val := range mapData {
		// Convert key to type K
		convertedKey, err := convertValue[K](key)
		if err != nil {
			return err
		}

		// Convert value to type V
		convertedVal, err := convertValue[V](val)
		if err != nil {
			return err
		}

		result[convertedKey] = convertedVal

		// Run callback for validation if provided
		if f.callback != nil {
			subSchema := NewSchema()
			f.callback(subSchema, convertedKey, convertedVal)
			err := subSchema.Apply(mapData)
			if err != nil {
				return fmt.Errorf("callback validation failed: %v", err)
			}
		}
	}

	*f.ptr = result
	f.wasAssigned = true

	return nil
}

// Validate validates the field value using all registered validators
func (f *MapField[K, V]) Validate(schema *Schema) error {
	return validateFieldValidators(f.Validators, *f.ptr, f.name, schema)
}

// AppendValidators implements ValidatorsAppender interface
func (f *MapField[K, V]) AppendValidators(validators []Validator) {
	f.Validators = append(f.Validators, validators...)
}

// SetCallback sets the callback function for configuring sub-schemas
func (f *MapField[K, V]) SetCallback(callback func(*Schema, K, V)) {
	f.callback = callback
}

// SetDefaultValue sets the default value for the field
func (f *MapField[K, V]) SetDefaultValue(defaultValue map[K]V) {
	f.defaultValue = defaultValue
	f.hasDefault = true
}

// Map creates a map field
func Map[K comparable, V any](name string, ptr *map[K]V, opts ...Option) Field {
	field := &MapField[K, V]{
		name: name,
		ptr:  ptr,
	}

	for _, opt := range opts {
		opt.Apply(field)
	}

	return field
}
