package poxxy

import (
	"fmt"
	"reflect"
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
