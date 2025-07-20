package poxxy

import (
	"database/sql/driver"
	"fmt"
	"reflect"
)

// PointerField represents a pointer field
type PointerField[T any] struct {
	name         string
	description  string
	ptr          **T
	Validators   []Validator
	callback     func(*Schema, *T)
	wasAssigned  bool // Track if a non-nil value was assigned
	defaultValue T
	hasDefault   bool
	transformers []Transformer[T]
}

// Name returns the field name
func (f *PointerField[T]) Name() string {
	return f.name
}

// Description returns the field description
func (f *PointerField[T]) Description() string {
	return f.description
}

// SetDescription sets the field description
func (f *PointerField[T]) SetDescription(description string) {
	f.description = description
}

// AddTransformer adds a transformer to the field
func (f *PointerField[T]) AddTransformer(transformer Transformer[T]) {
	f.transformers = append(f.transformers, transformer)
}

// SetDefaultValue sets the default value for the field
func (f *PointerField[T]) SetDefaultValue(defaultValue T) {
	f.defaultValue = defaultValue
	f.hasDefault = true
}

// Value returns the current value of the field
func (f *PointerField[T]) Value() interface{} {
	if f.ptr == nil {
		return nil
	}

	if !f.wasAssigned {
		return nil
	}

	// Use reflection to check if value implements driver.Valuer
	v := reflect.ValueOf(*f.ptr)
	if v.Type().Implements(reflect.TypeOf((*driver.Valuer)(nil)).Elem()) {
		if valuer, ok := v.Interface().(driver.Valuer); ok {
			value, err := valuer.Value()
			if err != nil {
				return nil
			}
			return value
		}
	}

	return *f.ptr
}

// Assign assigns a value to the field from the input data
func (f *PointerField[T]) Assign(data map[string]interface{}, schema *Schema) error {
	value, exists := data[f.name]
	if !exists {
		// Apply default value if available
		if f.hasDefault {
			instance := new(T)
			*instance = f.defaultValue
			*f.ptr = instance
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

	// Handle empty string for pointer fields - set to nil
	if str, ok := value.(string); ok && str == "" {
		*f.ptr = nil
		f.wasAssigned = true
		return nil
	}

	// Allocate new instance
	instance := new(T)
	*f.ptr = instance

	if f.callback != nil {
		structData, ok := value.(map[string]interface{})
		if !ok {
			return fmt.Errorf("expected object for struct pointer field")
		}
		subSchema := NewSchema()
		f.callback(subSchema, instance)
		f.wasAssigned = true
		return subSchema.Apply(structData)
	} else {
		converted, err := convertValue[T](value)
		if err != nil {
			return err
		}

		// Apply transformers
		transformed := converted
		for _, transformer := range f.transformers {
			transformed, err = transformer.Transform(transformed)
			if err != nil {
				return err
			}
		}

		**f.ptr = transformed
		f.wasAssigned = true
	}

	return nil
}

// Validate validates the field value using all registered validators
func (f *PointerField[T]) Validate(schema *Schema) error {
	if f.ptr == nil || *f.ptr == nil {
		return validateFieldValidators(f.Validators, nil, f.name, schema)
	}

	return validateFieldValidators(f.Validators, **f.ptr, f.name, schema)
}

// AppendValidators implements ValidatorsAppender interface
func (f *PointerField[T]) AppendValidators(validators []Validator) {
	f.Validators = append(f.Validators, validators...)
}

// SetCallback sets the callback function for configuring sub-schemas
func (f *PointerField[T]) SetCallback(callback func(*Schema, *T)) {
	f.callback = callback
}

// Pointer creates a pointer field
func Pointer[T any](name string, ptr **T, opts ...Option) Field {
	field := &PointerField[T]{
		name: name,
		ptr:  ptr,
	}

	for _, opt := range opts {
		opt.Apply(field)
	}

	return field
}
