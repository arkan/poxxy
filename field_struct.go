package poxxy

import (
	"fmt"
)

// StructField represents a struct field with callback
type StructField[T any] struct {
	name     string
	ptr      *T
	callback func(*Schema, *T)
}

func (f *StructField[T]) Name() string {
	return f.name
}

func (f *StructField[T]) Assign(data map[string]interface{}, schema *Schema) error {
	value, exists := data[f.name]
	if !exists {
		return nil
	}

	structData, ok := value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected object for struct field")
	}

	// Create a sub-schema and let the callback define it
	subSchema := NewSchema()
	f.callback(subSchema, f.ptr)

	// Assign and validate the struct data
	return subSchema.Apply(structData)
}

func (f *StructField[T]) Validate(schema *Schema) error {
	// Validation is done during assignment phase
	return nil
}

// Struct creates a struct field
func Struct[T any](name string, ptr *T, callback func(*Schema, *T)) Field {
	return &StructField[T]{
		name:     name,
		ptr:      ptr,
		callback: callback,
	}
}
