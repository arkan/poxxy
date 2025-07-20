package poxxy

import (
	"database/sql/driver"
	"reflect"
)

// ConvertField represents a field with type conversion
type ConvertField[From, To any] struct {
	name         string
	description  string
	ptr          *To
	convert      func(From) (*To, error)
	Validators   []Validator
	wasAssigned  bool // Track if a non-nil value was assigned
	defaultValue To
	hasDefault   bool
	transformers []Transformer[To]
}

// Name returns the field name
func (f *ConvertField[From, To]) Name() string {
	return f.name
}

// Description returns the field description
func (f *ConvertField[From, To]) Description() string {
	return f.description
}

// SetDescription sets the field description
func (f *ConvertField[From, To]) SetDescription(description string) {
	f.description = description
}

// AddTransformer adds a transformer to the field
func (f *ConvertField[From, To]) AddTransformer(transformer Transformer[To]) {
	f.transformers = append(f.transformers, transformer)
}

// SetDefaultValue sets the default value for the field
func (f *ConvertField[From, To]) SetDefaultValue(defaultValue To) {
	f.defaultValue = defaultValue
	f.hasDefault = true
}

// Value returns the current value of the field
func (f *ConvertField[From, To]) Value() interface{} {
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

	if _, ok := value.(string); ok && value.(string) == "" {
		f.wasAssigned = false
		return nil
	}

	// Convert from input type to From type
	fromValue, err := convertValue[From](value)
	if err != nil {
		return err
	}

	// Apply custom conversion
	converted, err := f.convert(fromValue)
	if err != nil {
		return err
	}

	// If converter returns nil, don't mutate the pointer
	if converted == nil {
		f.wasAssigned = false
		return nil
	}

	// Apply transformers
	transformed := *converted
	for _, transformer := range f.transformers {
		transformed, err = transformer.Transform(transformed)
		if err != nil {
			return err
		}
	}

	*f.ptr = transformed
	f.wasAssigned = true

	return nil
}

// Validate validates the field value using all registered validators
func (f *ConvertField[From, To]) Validate(schema *Schema) error {
	return validateFieldValidators(f.Validators, *f.ptr, f.name, schema)
}

// AppendValidators implements ValidatorsAppender interface
func (f *ConvertField[From, To]) AppendValidators(validators []Validator) {
	f.Validators = append(f.Validators, validators...)
}

// Convert creates a conversion field
func Convert[From, To any](name string, ptr *To, convert func(From) (*To, error), opts ...Option) Field {
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
func ConvertPointer[From, To any](name string, ptr **To, convert func(From) (*To, error), opts ...Option) Field {
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
	convert      func(From) (*To, error)
	Validators   []Validator
	wasAssigned  bool // Track if a non-nil value was assigned
	defaultValue To
	hasDefault   bool
	transformers []Transformer[To]
}

// Name returns the field name
func (f *ConvertPointerField[From, To]) Name() string {
	return f.name
}

// Description returns the field description
func (f *ConvertPointerField[From, To]) Description() string {
	return f.description
}

// SetDescription sets the field description
func (f *ConvertPointerField[From, To]) SetDescription(description string) {
	f.description = description
}

// AddTransformer adds a transformer to the field
func (f *ConvertPointerField[From, To]) AddTransformer(transformer Transformer[To]) {
	f.transformers = append(f.transformers, transformer)
}

// SetDefaultValue sets the default value for the field
func (f *ConvertPointerField[From, To]) SetDefaultValue(defaultValue To) {
	f.defaultValue = defaultValue
	f.hasDefault = true
}

// Value returns the current value of the field
func (f *ConvertPointerField[From, To]) Value() interface{} {
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
func (f *ConvertPointerField[From, To]) Assign(data map[string]interface{}, schema *Schema) error {
	value, exists := data[f.name]
	if !exists {
		// Apply default value if available
		if f.hasDefault {
			*f.ptr = &f.defaultValue
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

	if _, ok := value.(string); ok && value.(string) == "" {
		f.wasAssigned = false
		return nil
	}

	// Convert from input type to From type
	fromValue, err := convertValue[From](value)
	if err != nil {
		return err
	}

	// Apply custom conversion
	converted, err := f.convert(fromValue)
	if err != nil {
		return err
	}

	// If converter returns nil, don't mutate the pointer
	if converted == nil {
		f.wasAssigned = false
		return nil
	}

	// Apply transformers
	transformed := *converted
	for _, transformer := range f.transformers {
		transformed, err = transformer.Transform(transformed)
		if err != nil {
			return err
		}
	}

	*f.ptr = &transformed
	f.wasAssigned = true

	return nil
}

// Validate validates the field value using all registered validators
func (f *ConvertPointerField[From, To]) Validate(schema *Schema) error {
	if f.ptr == nil || *f.ptr == nil {
		return validateFieldValidators(f.Validators, nil, f.name, schema)
	}

	return validateFieldValidators(f.Validators, **f.ptr, f.name, schema)
}

// AppendValidators implements ValidatorsAppender interface
func (f *ConvertPointerField[From, To]) AppendValidators(validators []Validator) {
	f.Validators = append(f.Validators, validators...)
}
