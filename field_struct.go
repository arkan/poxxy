package poxxy

import (
	"fmt"
)

// StructField represents a struct field with callback
type StructField[T any] struct {
	name       string
	ptr        *T
	callback   func(*Schema, *T)
	Validators []Validator
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

	if f.callback == nil {
		return fmt.Errorf("callback is nil for field %s, did you forget to use WithSubSchema?", f.name)
	}

	// Create a sub-schema and let the callback define it
	subSchema := NewSchema()
	f.callback(subSchema, f.ptr)

	// Assign and validate the struct data
	return subSchema.Apply(structData)
}

func (f *StructField[T]) Validate(schema *Schema) error {
	return validateFieldValidators(f.Validators, *f.ptr, f.name, schema)
}

func (f *StructField[T]) SetCallback(callback func(*Schema, *T)) {
	f.callback = callback
}

// Struct creates a struct field
func Struct[T any](name string, ptr *T, opts ...Option) Field {
	field := &StructField[T]{
		name: name,
		ptr:  ptr,
	}

	for _, opt := range opts {
		opt.Apply(field)
	}

	return field
}
