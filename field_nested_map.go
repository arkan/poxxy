package poxxy

import (
	"fmt"
)

// NestedMapField represents a nested map field
type NestedMapField[K comparable, V any] struct {
	name        string
	description string
	ptr         *map[K]V
	callback    func(*Schema, K, *V)
}

func (f *NestedMapField[K, V]) Name() string {
	return f.name
}

func (f *NestedMapField[K, V]) Description() string {
	return f.description
}

func (f *NestedMapField[K, V]) SetDescription(description string) {
	f.description = description
}

func (f *NestedMapField[K, V]) Assign(data map[string]interface{}, schema *Schema) error {
	value, exists := data[f.name]
	if !exists {
		return nil
	}

	mapData, ok := value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected map for nested map field")
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
			valCopy := convertedVal
			f.callback(subSchema, convertedKey, &valCopy)
		}
	}

	*f.ptr = result
	return nil
}

func (f *NestedMapField[K, V]) Validate(schema *Schema) error {
	// Validation happens during assignment
	return nil
}

// NestedMap creates a nested map field
func NestedMap[K comparable, V any](name string, ptr *map[K]V, callback func(*Schema, K, *V)) Field {
	return &NestedMapField[K, V]{
		name:     name,
		ptr:      ptr,
		callback: callback,
	}
}
