package poxxy

import (
	"fmt"
)

// MapField represents a map field
type MapField[K comparable, V any] struct {
	name       string
	ptr        *map[K]V
	callback   func(*Schema, K, V)
	Validators []Validator
}

func (f *MapField[K, V]) Name() string {
	return f.name
}

func (f *MapField[K, V]) Assign(data map[string]interface{}, schema *Schema) error {
	value, exists := data[f.name]
	if !exists {
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
			return fmt.Errorf("key conversion failed: %v", err)
		}

		// Convert value to type V
		convertedVal, err := convertValue[V](val)
		if err != nil {
			return fmt.Errorf("value conversion failed: %v", err)
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
	return nil
}

func (f *MapField[K, V]) Validate(schema *Schema) error {
	return validateFieldValidators(f.Validators, *f.ptr, f.name, schema)
}

func (f *MapField[K, V]) SetCallback(callback func(*Schema, K, V)) {
	f.callback = callback
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
