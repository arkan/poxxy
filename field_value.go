package poxxy

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

// Name returns the field name
func (f *ValueField[T]) Name() string {
	return f.name
}

// Value returns the current value of the field
func (f *ValueField[T]) Value() interface{} {
	if f.ptr == nil {
		return nil
	}
	if !f.wasAssigned {
		return nil
	}
	return *f.ptr
}

// Description returns the field description
func (f *ValueField[T]) Description() string {
	return f.description
}

// SetDescription sets the field description
func (f *ValueField[T]) SetDescription(description string) {
	f.description = description
}

// AddTransformer adds a transformer to the field
func (f *ValueField[T]) AddTransformer(transformer Transformer[T]) {
	f.transformers = append(f.transformers, transformer)
}

// SetDefaultValue sets the default value for the field
func (f *ValueField[T]) SetDefaultValue(defaultValue T) {
	f.defaultValue = defaultValue
	f.hasDefault = true
}

// Assign assigns a value to the field from the input data
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

	*f.ptr = transformed
	f.wasAssigned = true
	return nil
}

// Validate validates the field value using all registered validators
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
