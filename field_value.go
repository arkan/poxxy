package poxxy

import "fmt"

// ValueField represents a basic value field
type ValueField[T any] struct {
	name         string
	description  string
	ptr          *T
	Validators   []Validator
	wasAssigned  bool // Track if a non-nil value was assigned
	defaultValue T
	hasDefault   bool
	transformers []Transformer[T]
}

func (f *ValueField[T]) Name() string {
	return f.name
}

func (f *ValueField[T]) Value() interface{} {
	if f.ptr == nil {
		return nil
	}
	if !f.wasAssigned {
		return nil
	}
	return *f.ptr
}

func (f *ValueField[T]) Description() string {
	return f.description
}

func (f *ValueField[T]) SetDescription(description string) {
	f.description = description
}

func (f *ValueField[T]) AddTransformer(transformer Transformer[T]) {
	f.transformers = append(f.transformers, transformer)
}

func (f *ValueField[T]) SetDefaultValue(defaultValue T) {
	f.defaultValue = defaultValue
	f.hasDefault = true
}

func (f *ValueField[T]) Assign(data map[string]interface{}, schema *Schema) error {
	value, exists := data[f.name]
	if !exists {
		// Apply default value if available
		if f.hasDefault {
			*f.ptr = f.defaultValue
			f.wasAssigned = true
			schema.SetFieldPresent(f.name)
		}
		return nil // Will be caught by Required validator if needed
	}
	schema.SetFieldPresent(f.name)

	if value == nil {
		f.wasAssigned = false
		return nil
	}

	// Handle empty string for value fields
	if str, ok := value.(string); ok && str == "" {
		// For empty strings, set to zero value
		var zero T
		*f.ptr = zero
		f.wasAssigned = true
		return nil
	}

	// Type conversion
	converted, err := convertValue[T](value)
	if err != nil {
		return fmt.Errorf("type conversion failed: %v", err)
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

func (f *ValueField[T]) Validate(schema *Schema) error {
	return validateFieldValidators(f.Validators, *f.ptr, f.name, schema)
}

// AppendValidators implements ValidatorsAppender interface
func (f *ValueField[T]) AppendValidators(validators []Validator) {
	f.Validators = append(f.Validators, validators...)
}

// Value creates a value field
func Value[T any](name string, ptr *T, opts ...Option) Field {
	field := &ValueField[T]{
		name: name,
		ptr:  ptr,
	}

	for _, opt := range opts {
		opt.Apply(field)
	}

	return field
}
