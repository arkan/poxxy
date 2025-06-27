package poxxy

import (
	"fmt"
)

// ValueWithoutAssignField represents a field that validates a direct value
type ValueWithoutAssignField[T any] struct {
	name        string
	description string
	value       interface{}
	Validators  []Validator
	wasAssigned bool // Track if a non-nil value was assigned
}

func (f *ValueWithoutAssignField[T]) Name() string {
	return f.name
}

func (f *ValueWithoutAssignField[T]) Description() string {
	return f.description
}

func (f *ValueWithoutAssignField[T]) SetDescription(description string) {
	f.description = description
}

func (f *ValueWithoutAssignField[T]) Value() interface{} {
	if f.value == nil {
		return nil
	}
	if !f.wasAssigned {
		return nil
	}
	return f.value
}

func (f *ValueWithoutAssignField[T]) Assign(data map[string]interface{}, schema *Schema) error {
	value, exists := data[f.name]
	if !exists {
		return nil
	}

	schema.SetFieldPresent(f.name)

	if value == nil {
		f.wasAssigned = false
		return nil
	}

	converted, err := convertValue[T](value)
	if err != nil {
		return fmt.Errorf("type conversion failed: %v", err)
	}

	f.value = converted
	f.wasAssigned = true
	return nil
}

func (f *ValueWithoutAssignField[T]) Validate(schema *Schema) error {
	return validateFieldValidators(f.Validators, f.value, f.name, schema)
}

// ValueWithoutAssign validates a direct value (used in map validation)
func ValueWithoutAssign[T any](name string, opts ...Option) Field {
	field := &ValueWithoutAssignField[T]{
		name: name,
	}

	for _, opt := range opts {
		opt.Apply(field)
	}

	return field
}
