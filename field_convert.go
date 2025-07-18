package poxxy

import (
	"fmt"
)

// ConvertField represents a field with type conversion
type ConvertField[From, To any] struct {
	name         string
	description  string
	ptr          *To
	convert      func(From) (To, error)
	Validators   []Validator
	wasAssigned  bool // Track if a non-nil value was assigned
	defaultValue To
	hasDefault   bool
	transformers []Transformer[To]
}

func (f *ConvertField[From, To]) Name() string {
	return f.name
}

func (f *ConvertField[From, To]) Description() string {
	return f.description
}

func (f *ConvertField[From, To]) SetDescription(description string) {
	f.description = description
}

func (f *ConvertField[From, To]) AddTransformer(transformer Transformer[To]) {
	f.transformers = append(f.transformers, transformer)
}

func (f *ConvertField[From, To]) SetDefaultValue(defaultValue To) {
	f.defaultValue = defaultValue
	f.hasDefault = true
}

func (f *ConvertField[From, To]) Value() interface{} {
	if f.ptr == nil {
		return nil
	}
	if !f.wasAssigned {
		return nil
	}
	return *f.ptr
}

func (f *ConvertField[From, To]) Assign(data map[string]interface{}, schema *Schema) error {
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

	// Convert from input type to From type
	fromValue, err := convertValue[From](value)
	if err != nil {
		return fmt.Errorf("type conversion failed: %v", err)
	}

	// Apply custom conversion
	converted, err := f.convert(fromValue)
	if err != nil {
		return fmt.Errorf("conversion failed: %v", err)
	}

	// Apply transformers
	transformed := converted
	for _, transformer := range f.transformers {
		transformed, err = transformer.Transform(transformed)
		if err != nil {
			return fmt.Errorf("transformer failed: %v", err)
		}
	}

	*f.ptr = transformed
	f.wasAssigned = true
	return nil
}

func (f *ConvertField[From, To]) Validate(schema *Schema) error {
	return validateFieldValidators(f.Validators, *f.ptr, f.name, schema)
}

// Convert creates a conversion field
func Convert[From, To any](name string, ptr *To, convert func(From) (To, error), opts ...Option) Field {
	field := &ConvertField[From, To]{
		name:    name,
		ptr:     ptr,
		convert: convert,
	}

	for _, opt := range opts {
		opt.Apply(field)
	}

	return field
}

// ConvertPointer creates a conversion field for pointer types
func ConvertPointer[From, To any](name string, ptr **To, convert func(From) (To, error), opts ...Option) Field {
	field := &ConvertPointerField[From, To]{
		name:    name,
		ptr:     ptr,
		convert: convert,
	}

	for _, opt := range opts {
		opt.Apply(field)
	}

	return field
}

// ConvertPointerField represents a field with type conversion to pointer
type ConvertPointerField[From, To any] struct {
	name         string
	description  string
	ptr          **To
	convert      func(From) (To, error)
	Validators   []Validator
	wasAssigned  bool // Track if a non-nil value was assigned
	defaultValue To
	hasDefault   bool
	transformers []Transformer[To]
}

func (f *ConvertPointerField[From, To]) Name() string {
	return f.name
}

func (f *ConvertPointerField[From, To]) Description() string {
	return f.description
}

func (f *ConvertPointerField[From, To]) SetDescription(description string) {
	f.description = description
}

func (f *ConvertPointerField[From, To]) AddTransformer(transformer Transformer[To]) {
	f.transformers = append(f.transformers, transformer)
}

func (f *ConvertPointerField[From, To]) SetDefaultValue(defaultValue To) {
	f.defaultValue = defaultValue
	f.hasDefault = true
}

func (f *ConvertPointerField[From, To]) Value() interface{} {
	if f.ptr == nil {
		return nil
	}
	if !f.wasAssigned {
		return nil
	}
	return *f.ptr
}

func (f *ConvertPointerField[From, To]) Assign(data map[string]interface{}, schema *Schema) error {
	value, exists := data[f.name]
	if !exists {
		// Apply default value if available
		if f.hasDefault {
			instance := new(To)
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

	// Convert from input type to From type
	fromValue, err := convertValue[From](value)
	if err != nil {
		return fmt.Errorf("type conversion failed: %v", err)
	}

	// Apply custom conversion
	converted, err := f.convert(fromValue)
	if err != nil {
		return fmt.Errorf("conversion failed: %v", err)
	}

	// Apply transformers
	transformed := converted
	for _, transformer := range f.transformers {
		transformed, err = transformer.Transform(transformed)
		if err != nil {
			return fmt.Errorf("transformer failed: %v", err)
		}
	}

	// Allocate new instance and assign
	instance := new(To)
	*instance = transformed
	*f.ptr = instance
	f.wasAssigned = true
	return nil
}

func (f *ConvertPointerField[From, To]) Validate(schema *Schema) error {
	if *f.ptr == nil {
		// Pointer is nil - skip validation unless required
		return validateFieldValidators(f.Validators, nil, f.name, schema)
	}

	// Validate the pointed value
	return validateFieldValidators(f.Validators, **f.ptr, f.name, schema)
}
