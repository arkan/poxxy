package poxxy

import (
	"database/sql/driver"
	"reflect"
)

// ValueWithoutAssignField represents a field that validates a direct value
type ValueWithoutAssignField[T any] struct {
	name        string
	description string
	value       interface{}
	Validators  []Validator
	wasAssigned bool // Track if a non-nil value was assigned
}

// Name returns the field name
func (f *ValueWithoutAssignField[T]) Name() string {
	return f.name
}

// Description returns the field description
func (f *ValueWithoutAssignField[T]) Description() string {
	return f.description
}

// SetDescription sets the field description
func (f *ValueWithoutAssignField[T]) SetDescription(description string) {
	f.description = description
}

// Value returns the current value of the field
func (f *ValueWithoutAssignField[T]) Value() interface{} {
	if f.value == nil {
		return nil
	}
	if !f.wasAssigned {
		return nil
	}

	// Use reflection to check if value implements driver.Valuer
	v := reflect.ValueOf(f.value)
	if v.Type().Implements(reflect.TypeOf((*driver.Valuer)(nil)).Elem()) {
		if valuer, ok := v.Interface().(driver.Valuer); ok {
			value, err := valuer.Value()
			if err != nil {
				return nil
			}
			return value
		}
	}

	return f.value
}

// Assign assigns a value to the field from the input data
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
		return err
	}

	f.value = converted
	f.wasAssigned = true
	return nil
}

// Validate validates the field value using all registered validators
func (f *ValueWithoutAssignField[T]) Validate(schema *Schema) error {
	return validateFieldValidators(f.Validators, f.value, f.name, schema)
}

// AppendValidators implements ValidatorsAppender interface
func (f *ValueWithoutAssignField[T]) AppendValidators(validators []Validator) {
	f.Validators = append(f.Validators, validators...)
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
