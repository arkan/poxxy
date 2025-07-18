package poxxy

import (
	"fmt"
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

func (f *PointerField[T]) Name() string {
	return f.name
}

func (f *PointerField[T]) Description() string {
	return f.description
}

func (f *PointerField[T]) SetDescription(description string) {
	f.description = description
}

func (f *PointerField[T]) AddTransformer(transformer Transformer[T]) {
	f.transformers = append(f.transformers, transformer)
}

func (f *PointerField[T]) SetDefaultValue(defaultValue T) {
	f.defaultValue = defaultValue
	f.hasDefault = true
}

func (f *PointerField[T]) Value() interface{} {
	if f.ptr == nil {
		return nil
	}
	if !f.wasAssigned {
		return nil
	}
	return *f.ptr
}

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
			return fmt.Errorf("pointer field conversion failed: %v", err)
		}

		// Apply transformers
		transformed := converted
		for _, transformer := range f.transformers {
			transformed, err = transformer.Transform(transformed)
			if err != nil {
				return fmt.Errorf("transformer failed: %v", err)
			}
		}

		**f.ptr = transformed
		f.wasAssigned = true
	}

	return nil
}

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
