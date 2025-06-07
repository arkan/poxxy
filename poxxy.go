package poxxy

import (
	"fmt"
	"net/http"
	"reflect"
	"strconv"
)

// Validator represents a validation function
type Validator interface {
	Validate(value interface{}, fieldName string) error
	WithMessage(msg string) Validator
}

// ValidatorFunc is a function that implements Validator
type ValidatorFunc struct {
	fn  func(interface{}, string) error
	msg string
}

func (v ValidatorFunc) Validate(value interface{}, fieldName string) error {
	err := v.fn(value, fieldName)
	if err != nil && v.msg != "" {
		return fmt.Errorf("%s", v.msg)
	}
	return err
}

func (v ValidatorFunc) WithMessage(msg string) Validator {
	return ValidatorFunc{fn: v.fn, msg: msg}
}

// FieldError represents a validation error for a specific field
type FieldError struct {
	Field string
	Error error
}

// Errors represents multiple validation errors
type Errors []FieldError

func (e Errors) Error() string {
	var msgs []string
	for _, err := range e {
		msgs = append(msgs, fmt.Sprintf("%s: %v", err.Field, err.Error))
	}
	// Manual join instead of using strings.Join
	if len(msgs) == 0 {
		return ""
	}
	result := msgs[0]
	for i := 1; i < len(msgs); i++ {
		result += "; " + msgs[i]
	}
	return result
}

// Field represents a field definition in a schema
type Field interface {
	Name() string
	Assign(data map[string]interface{}, schema *Schema) error
	Validate(schema *Schema) error
}

// Schema represents a validation schema
type Schema struct {
	fields        []Field
	data          map[string]interface{}
	presentFields map[string]bool // Track which fields were present in input data
}

// NewSchema creates a new schema with the given fields
func NewSchema(fields ...Field) *Schema {
	return &Schema{
		fields:        fields,
		presentFields: make(map[string]bool),
	}
}

// ApplyHTTPRequest assigns data from an HTTP request to a schema
func (s *Schema) ApplyHTTPRequest(r *http.Request) error {
	if err := r.ParseForm(); err != nil {
		return fmt.Errorf("failed to parse form: %w", err)
	}

	form := make(map[string]interface{})

	for key, values := range r.Form {
		form[key] = values[0]
	}

	return s.Apply(form)
}

// Apply assigns data to variables and validates them
func (s *Schema) Apply(data map[string]interface{}) error {
	s.data = data
	s.presentFields = make(map[string]bool)

	// Track which top-level fields are present
	for key := range data {
		s.presentFields[key] = true
	}

	var errors Errors

	// First pass: assign values
	for _, field := range s.fields {
		if err := field.Assign(data, s); err != nil {
			errors = append(errors, FieldError{Field: field.Name(), Error: err})
		}
	}

	// If there are any errors, return them
	if len(errors) > 0 {
		return errors
	}

	// Second pass: validate
	for _, field := range s.fields {
		if err := field.Validate(s); err != nil {
			errors = append(errors, FieldError{Field: field.Name(), Error: err})
		}
	}

	// If there are any errors, return them
	if len(errors) > 0 {
		return errors
	}

	return nil
}

// validateFieldValidators is a helper function to validate a list of validators, handling RequiredValidator specially
func validateFieldValidators(validators []Validator, value interface{}, fieldName string, schema *Schema) error {
	for _, validator := range validators {
		// Handle RequiredValidator specially - it needs schema context
		if reqValidator, ok := validator.(RequiredValidator); ok {
			if err := reqValidator.ValidateWithSchema(schema, fieldName); err != nil {
				return err
			}
		} else {
			if err := validator.Validate(value, fieldName); err != nil {
				return err
			}
		}
	}
	return nil
}

// IsFieldPresent checks if a field was present in the input data
func (s *Schema) IsFieldPresent(fieldName string) bool {
	return s.presentFields[fieldName]
}

// WithSchema helper function to add fields to a schema
func WithSchema(schema *Schema, field Field) {
	schema.fields = append(schema.fields, field)
}

// ValueField represents a basic value field
type ValueField[T any] struct {
	name       string
	ptr        *T
	Validators []Validator
}

func (f *ValueField[T]) Name() string {
	return f.name
}

func (f *ValueField[T]) Assign(data map[string]interface{}, schema *Schema) error {
	value, exists := data[f.name]
	if !exists {
		return nil // Will be caught by Required validator if needed
	}

	// Type conversion
	converted, err := convertValue[T](value)
	if err != nil {
		return fmt.Errorf("type conversion failed: %v", err)
	}

	*f.ptr = converted
	return nil
}

func (f *ValueField[T]) Validate(schema *Schema) error {
	return validateFieldValidators(f.Validators, *f.ptr, f.name, schema)
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

// convertValue converts interface{} to type T
func convertValue[T any](value interface{}) (T, error) {
	var zero T

	// Direct type assertion first
	if v, ok := value.(T); ok {
		return v, nil
	}

	// Handle string conversions
	targetType := reflect.TypeOf(zero)
	sourceValue := reflect.ValueOf(value)

	if targetType == sourceValue.Type() {
		return value.(T), nil
	}

	// Convert based on target type
	switch targetType.Kind() {
	case reflect.String:
		str := fmt.Sprintf("%v", value)
		return any(str).(T), nil
	case reflect.Int, reflect.Int64:
		switch v := value.(type) {
		case int:
			return any(v).(T), nil
		case int64:
			return any(v).(T), nil
		case float64:
			return any(int(v)).(T), nil
		case string:
			if i, err := strconv.Atoi(v); err == nil {
				return any(i).(T), nil
			}
		}
	case reflect.Float64:
		switch v := value.(type) {
		case float64:
			return any(v).(T), nil
		case int:
			return any(float64(v)).(T), nil
		case int64:
			return any(float64(v)).(T), nil
		case string:
			if f, err := strconv.ParseFloat(v, 64); err == nil {
				return any(f).(T), nil
			}
		}
	case reflect.Bool:
		if b, ok := value.(bool); ok {
			return any(b).(T), nil
		}
	}

	return zero, fmt.Errorf("cannot convert %T to %T", value, zero)
}

// Option represents a configuration option
type Option interface {
	Apply(interface{})
}

// ValidatorsOption holds validators
type ValidatorsOption struct {
	validators []Validator
}

func (o ValidatorsOption) Apply(field interface{}) {
	// Use type switching to handle validators for different field types
	switch f := field.(type) {
	case *ValueField[string]:
		f.Validators = append(f.Validators, o.validators...)
	case *ValueField[int]:
		f.Validators = append(f.Validators, o.validators...)
	case *ValueField[bool]:
		f.Validators = append(f.Validators, o.validators...)
	case *ValueField[float64]:
		f.Validators = append(f.Validators, o.validators...)
	case *ValueField[[]string]:
		f.Validators = append(f.Validators, o.validators...)
	case *ValueField[[4]int]:
		f.Validators = append(f.Validators, o.validators...)
	case *ValueField[map[string]string]:
		f.Validators = append(f.Validators, o.validators...)
	case *ArrayField[string]:
		f.Validators = append(f.Validators, o.validators...)
	case *ArrayField[int]:
		f.Validators = append(f.Validators, o.validators...)
	case *SliceField[string]:
		f.Validators = append(f.Validators, o.validators...)
	case *SliceField[int]:
		f.Validators = append(f.Validators, o.validators...)
	case *SliceField[float64]:
		f.Validators = append(f.Validators, o.validators...)
	case *ValueFromField[string]:
		f.Validators = append(f.Validators, o.validators...)
	case *ValueFromField[int]:
		f.Validators = append(f.Validators, o.validators...)
	case *ValueFromField[bool]:
		f.Validators = append(f.Validators, o.validators...)
	default:
		// Fallback to reflection for types we haven't explicitly handled
		fieldValue := reflect.ValueOf(field)
		if fieldValue.Kind() == reflect.Ptr {
			fieldValue = fieldValue.Elem()
		}

		validatorsField := fieldValue.FieldByName("Validators")
		if validatorsField.IsValid() && validatorsField.CanSet() {
			// Handle the validators field safely
			if validatorsField.Type() == reflect.TypeOf([]Validator{}) {
				currentValidators := validatorsField.Interface().([]Validator)
				newValidators := append(currentValidators, o.validators...)
				validatorsField.Set(reflect.ValueOf(newValidators))
			}
		}
	}
}

// WithValidators creates a validators option
func WithValidators(validators ...Validator) Option {
	return ValidatorsOption{validators: validators}
}
