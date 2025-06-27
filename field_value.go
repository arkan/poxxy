package poxxy

import "fmt"

// ValueField represents a basic value field
type ValueField[T any] struct {
	name        string
	description string
	ptr         *T
	Validators  []Validator
	wasAssigned bool // Track if a non-nil value was assigned
}

func (f *ValueField[T]) Name() string {
	return f.name
}

func (f *ValueField[T]) Value() interface{} {
	if f.ptr == nil {
		return nil
	}
	if !f.wasAssigned {
		return nil
	}
	return *f.ptr
}

func (f *ValueField[T]) Description() string {
	return f.description
}

func (f *ValueField[T]) SetDescription(description string) {
	f.description = description
}

func (f *ValueField[T]) Assign(data map[string]interface{}, schema *Schema) error {
	value, exists := data[f.name]
	if !exists {
		return nil // Will be caught by Required validator if needed
	}
	schema.SetFieldPresent(f.name)

	if value == nil {
		f.wasAssigned = false
		return nil
	}

	// Type conversion
	converted, err := convertValue[T](value)
	if err != nil {
		return fmt.Errorf("type conversion failed: %v", err)
	}

	*f.ptr = converted
	f.wasAssigned = true
	return nil
}

func (f *ValueField[T]) Validate(schema *Schema) error {
	return validateFieldValidators(f.Validators, *f.ptr, f.name, schema)
}

// Value creates a value field
func Value[T any](name string, ptr *T, opts ...Option) Field {
	field := &ValueField[T]{
		name: name,
		ptr:  ptr,
	}

	for _, opt := range opts {
		opt.Apply(field)
	}

	return field
}
