package poxxy

import (
	"fmt"
)

// StructField represents a struct field with callback
type StructField[T any] struct {
	name         string
	description  string
	ptr          *T
	callback     func(*Schema, *T)
	Validators   []Validator
	wasAssigned  bool // Track if a non-nil value was assigned
	defaultValue T
	hasDefault   bool
}

// Name returns the field name
func (f *StructField[T]) Name() string {
	return f.name
}

// Value returns the current value of the field
func (f *StructField[T]) Value() interface{} {
	if f.ptr == nil {
		return nil
	}

	if !f.wasAssigned {
		return nil
	}

	return *f.ptr
}

// Description returns the field description
func (f *StructField[T]) Description() string {
	return f.description
}

// SetDescription sets the field description
func (f *StructField[T]) SetDescription(description string) {
	f.description = description
}

// Assign assigns a value to the field from the input data
func (f *StructField[T]) Assign(data map[string]interface{}, schema *Schema) error {
	value, exists := data[f.name]
	if !exists {
		// Apply default value if available
		if f.hasDefault {
			*f.ptr = f.defaultValue
			f.wasAssigned = true
			schema.SetFieldPresent(f.name)
		}

		return nil
	}

	schema.SetFieldPresent(f.name)

	if value == nil {
		f.wasAssigned = false
		return nil
	}

	structData, ok := value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected object for struct field")
	}

	if f.callback == nil {
		return fmt.Errorf("callback is nil for field %s, did you forget to use WithSubSchema?", f.name)
	}

	subSchema := NewSchema()
	f.callback(subSchema, f.ptr)
	f.wasAssigned = true

	return subSchema.Apply(structData)
}

// Validate validates the field value using all registered validators
func (f *StructField[T]) Validate(schema *Schema) error {
	return validateFieldValidators(f.Validators, *f.ptr, f.name, schema)
}

// AppendValidators implements ValidatorsAppender interface
func (f *StructField[T]) AppendValidators(validators []Validator) {
	f.Validators = append(f.Validators, validators...)
}

// SetCallback sets the callback function for configuring sub-schemas
func (f *StructField[T]) SetCallback(callback func(*Schema, *T)) {
	f.callback = callback
}

// SetDefaultValue sets the default value for the field
func (f *StructField[T]) SetDefaultValue(defaultValue T) {
	f.defaultValue = defaultValue
	f.hasDefault = true
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
