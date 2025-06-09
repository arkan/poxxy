package poxxy

import (
	"fmt"
)

// ValueFromField represents a field that validates a direct value
type ValueFromField[T any] struct {
	name       string
	value      interface{}
	Validators []Validator
}

func (f *ValueFromField[T]) Name() string {
	return f.name
}

func (f *ValueFromField[T]) Assign(data map[string]interface{}, schema *Schema) error {
	value, exists := data[f.name]
	if !exists {
		return nil
	}

	converted, err := convertValue[T](value)
	if err != nil {
		return fmt.Errorf("type conversion failed: %v", err)
	}

	f.value = converted

	return nil
}

func (f *ValueFromField[T]) Validate(schema *Schema) error {
	return validateFieldValidators(f.Validators, f.value, f.name, schema)
}

// V validates a direct value (used in map validation)
func V[T any](name string, opts ...Option) Field {
	field := &ValueFromField[T]{
		name: name,
	}

	for _, opt := range opts {
		opt.Apply(field)
	}

	return field
}
