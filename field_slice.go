package poxxy

import (
	"fmt"
	"reflect"
)

// SliceField represents a slice field where each element is a struct
type SliceField[T any] struct {
	name        string
	description string
	ptr         *[]T
	callback    func(*Schema, *T)
	Validators  []Validator
}

func (f *SliceField[T]) Name() string {
	return f.name
}

func (f *SliceField[T]) Description() string {
	return f.description
}

func (f *SliceField[T]) SetDescription(description string) {
	f.description = description
}

func (f *SliceField[T]) Assign(data map[string]interface{}, schema *Schema) error {
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
		switch v := item.(type) {
		// If the item is a map, we need to create a new instance for this element
		// THis is the case when we want to map the element to a struct.
		case map[string]interface{}:
			// Create a new instance for this element
			var element T

			// Create a sub-schema for this element
			subSchema := NewSchema()

			// Apply the callback to define the schema for this element
			if f.callback != nil {
				f.callback(subSchema, &element)
			}

			// Assign and validate this element
			if err := subSchema.Apply(v); err != nil {
				return fmt.Errorf("element %d: %v", i, err)
			}

			result[i] = element
		default:
			converted, err := convertValue[T](v)
			if err != nil {
				return fmt.Errorf("element %d: %v", i, err)
			}
			result[i] = converted
		}
	}

	*f.ptr = result
	return nil
}

func (f *SliceField[T]) Validate(schema *Schema) error {
	return validateFieldValidators(f.Validators, *f.ptr, f.name, schema)
}

func (f *SliceField[T]) SetCallback(callback func(*Schema, *T)) {
	f.callback = callback
}

// Slice creates a slice field.
func Slice[T any](name string, ptr *[]T, opts ...Option) Field {
	field := &SliceField[T]{
		name: name,
		ptr:  ptr,
	}

	for _, opt := range opts {
		opt.Apply(field)
	}

	return field
}
