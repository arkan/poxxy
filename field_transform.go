package poxxy

import (
	"fmt"
)

// TransformField represents a field with type transformation
type TransformField[From, To any] struct {
	name        string
	description string
	ptr         *To
	transform   func(From) (To, error)
	Validators  []Validator
	wasAssigned bool // Track if a non-nil value was assigned
}

func (f *TransformField[From, To]) Name() string {
	return f.name
}

func (f *TransformField[From, To]) Description() string {
	return f.description
}

func (f *TransformField[From, To]) SetDescription(description string) {
	f.description = description
}

func (f *TransformField[From, To]) Value() interface{} {
	if f.ptr == nil {
		return nil
	}
	if !f.wasAssigned {
		return nil
	}
	return *f.ptr
}

func (f *TransformField[From, To]) Assign(data map[string]interface{}, schema *Schema) error {
	value, exists := data[f.name]
	if !exists {
		return nil
	}

	schema.SetFieldPresent(f.name)

	if value == nil {
		f.wasAssigned = false
		return nil
	}

	fromValue, err := convertValue[From](value)
	if err != nil {
		return fmt.Errorf("type conversion failed: %v", err)
	}

	toValue, err := f.transform(fromValue)
	if err != nil {
		return fmt.Errorf("transform failed: %v", err)
	}

	*f.ptr = toValue
	f.wasAssigned = true
	return nil
}

func (f *TransformField[From, To]) Validate(schema *Schema) error {
	return validateFieldValidators(f.Validators, *f.ptr, f.name, schema)
}

// Transform creates a transformation field
func Transform[From, To any](name string, ptr *To, transform func(From) (To, error), opts ...Option) Field {
	field := &TransformField[From, To]{
		name:      name,
		ptr:       ptr,
		transform: transform,
	}

	for _, opt := range opts {
		opt.Apply(field)
	}

	return field
}
