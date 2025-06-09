package poxxy

import (
	"fmt"
	"reflect"
)

// SliceField represents a slice field
type SliceField[T any] struct {
	name       string
	ptr        *[]T
	Validators []Validator
}

func (f *SliceField[T]) Name() string {
	return f.name
}

func (f *SliceField[T]) Assign(data map[string]interface{}, schema *Schema) error {
	value, exists := data[f.name]
	if !exists {
		return nil
	}

	sourceValue := reflect.ValueOf(value)
	if sourceValue.Kind() != reflect.Slice && sourceValue.Kind() != reflect.Array {
		return fmt.Errorf("source value must be slice or array")
	}

	// Create new slice
	result := make([]T, sourceValue.Len())

	// Convert elements
	for i := 0; i < sourceValue.Len(); i++ {
		srcElem := sourceValue.Index(i).Interface()
		converted, err := convertValue[T](srcElem)
		if err != nil {
			return fmt.Errorf("element %d: %v", i, err)
		}
		result[i] = converted
	}

	*f.ptr = result
	return nil
}

func (f *SliceField[T]) Validate(schema *Schema) error {
	for _, validator := range f.Validators {
		if err := validator.Validate(*f.ptr, f.name); err != nil {
			return err
		}
	}
	return nil
}

// Slice creates a slice field
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
