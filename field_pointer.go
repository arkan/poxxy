package poxxy

import (
	"fmt"
)

// PointerField represents a pointer field
type PointerField[T any] struct {
	name       string
	ptr        **T
	Validators []Validator
	callback   func(*Schema, *T)
}

func (f *PointerField[T]) Name() string {
	return f.name
}

func (f *PointerField[T]) Assign(data map[string]interface{}, schema *Schema) error {
	value, exists := data[f.name]
	if !exists {
		// Optional field - leave as nil
		return nil
	}

	// Allocate new instance
	instance := new(T)
	*f.ptr = instance

	if f.callback != nil {
		// Handle struct pointer with callback
		structData, ok := value.(map[string]interface{})
		if !ok {
			return fmt.Errorf("expected object for struct pointer field")
		}

		subSchema := NewSchema()
		f.callback(subSchema, instance)
		return subSchema.Apply(structData)
	} else {
		// Handle simple pointer
		converted, err := convertValue[T](value)
		if err != nil {
			return fmt.Errorf("pointer field conversion failed: %v", err)
		}
		**f.ptr = converted
	}

	return nil
}

func (f *PointerField[T]) Validate(schema *Schema) error {
	if *f.ptr == nil {
		// Pointer is nil - skip validation unless required
		return validateFieldValidators(f.Validators, nil, f.name, schema)
	}

	// Validate the pointed value
	return validateFieldValidators(f.Validators, **f.ptr, f.name, schema)
}

// Pointer creates a pointer field
func Pointer[T any](name string, ptr **T, opts ...interface{}) Field {
	var validators []Validator
	var callback func(*Schema, *T)

	for _, opt := range opts {
		switch o := opt.(type) {
		case Option:
			if validatorOpt, ok := o.(ValidatorsOption); ok {
				validators = append(validators, validatorOpt.validators...)
			}
		case func(*Schema, *T):
			callback = o
		}
	}

	field := &PointerField[T]{
		name:       name,
		ptr:        ptr,
		Validators: validators,
		callback:   callback,
	}

	return field
}
