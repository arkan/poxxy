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
	// ValueFrom doesn't assign, it validates existing values
	return nil
}

func (f *ValueFromField[T]) Validate(schema *Schema) error {
	converted, err := convertValue[T](f.value)
	if err != nil {
		return fmt.Errorf("type conversion failed: %v", err)
	}

	for _, validator := range f.Validators {
		if err := validator.Validate(converted, f.name); err != nil {
			return err
		}
	}
	return nil
}

// ValueFrom validates a direct value (used in nested map validation)
func ValueFrom[T any](name string, value interface{}, opts ...Option) Field {
	field := &ValueFromField[T]{
		name:  name,
		value: value,
	}

	for _, opt := range opts {
		opt.Apply(field)
	}

	return field
}
