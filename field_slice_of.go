package poxxy

import (
	"fmt"
	"reflect"
)

// SliceOfField represents a slice field where each element is a struct
type SliceOfField[T any] struct {
	name       string
	ptr        *[]T
	callback   func(*Schema, *T)
	Validators []Validator
}

func (f *SliceOfField[T]) Name() string {
	return f.name
}

func (f *SliceOfField[T]) Assign(data map[string]interface{}, schema *Schema) error {
	value, exists := data[f.name]
	if !exists {
		return nil
	}

	// Convert to slice of interface{} - handle different slice types
	var slice []interface{}

	switch v := value.(type) {
	case []interface{}:
		slice = v
	case []map[string]interface{}:
		// Convert []map[string]interface{} to []interface{}
		slice = make([]interface{}, len(v))
		for i, item := range v {
			slice[i] = item
		}
	default:
		// Try to use reflection to handle other slice types
		rValue := reflect.ValueOf(value)
		if rValue.Kind() != reflect.Slice {
			return fmt.Errorf("expected slice, got %T", value)
		}

		slice = make([]interface{}, rValue.Len())
		for i := 0; i < rValue.Len(); i++ {
			slice[i] = rValue.Index(i).Interface()
		}
	}

	// Create result slice
	result := make([]T, len(slice))

	// Process each element
	for i, item := range slice {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			return fmt.Errorf("element %d: expected map, got %T", i, item)
		}

		// Create a new instance for this element
		var element T

		// Create a sub-schema for this element
		subSchema := NewSchema()

		// Apply the callback to define the schema for this element
		if f.callback != nil {
			f.callback(subSchema, &element)
		}

		// Assign and validate this element
		if err := subSchema.Apply(itemMap); err != nil {
			return fmt.Errorf("element %d: %v", i, err)
		}

		result[i] = element
	}

	*f.ptr = result
	return nil
}

func (f *SliceOfField[T]) Validate(schema *Schema) error {
	return ValidateFieldValidators(f.Validators, *f.ptr, f.name, schema)
}

// SliceOf creates a slice field for structs with element-wise schema definition
func SliceOf[T any](name string, ptr *[]T, callback func(*Schema, *T), opts ...Option) Field {
	field := &SliceOfField[T]{
		name:     name,
		ptr:      ptr,
		callback: callback,
	}

	for _, opt := range opts {
		opt.Apply(field)
	}

	return field
}
