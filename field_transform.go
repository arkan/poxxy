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

func (f *TransformField[From, To]) Assign(data map[string]interface{}, schema *Schema) error {
	value, exists := data[f.name]
	if !exists {
		return nil
	}

	// Convert to From type first
	fromValue, err := convertValue[From](value)
	if err != nil {
		return fmt.Errorf("transform source conversion failed: %v", err)
	}

	// Apply transformation
	toValue, err := f.transform(fromValue)
	if err != nil {
		return fmt.Errorf("transformation failed: %v", err)
	}

	*f.ptr = toValue
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
