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

// ValidatorFn is a function that implements Validator
type ValidatorFn struct {
	fn  func(interface{}, string) error
	msg string
}

func (v ValidatorFn) Validate(value interface{}, fieldName string) error {
	err := v.fn(value, fieldName)
	if err != nil && v.msg != "" {
		return fmt.Errorf("%s", v.msg)
	}
	return err
}

func (v ValidatorFn) WithMessage(msg string) Validator {
	return ValidatorFn{fn: v.fn, msg: msg}
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

// ValidatorsAppender is an interface for fields that can append validators
type ValidatorsAppender interface {
	AppendValidators(validators []Validator)
}

// ValidatorsOption holds validators
type ValidatorsOption struct {
	validators []Validator
}

func (o ValidatorsOption) Apply(field interface{}) {
	// Try to use the interface first
	if appender, ok := field.(ValidatorsAppender); ok {
		appender.AppendValidators(o.validators)
		return
	}

	// Fallback to reflection for types that don't implement the interface
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

// WithValidators creates a validators option
func WithValidators(validators ...Validator) Option {
	return ValidatorsOption{validators: validators}
}

// DefaultValueSetter is an interface for fields that can set default values
type DefaultValueSetter[T any] interface {
	SetDefaultValue(defaultValue T)
}

// DefaultOption holds a default value
type DefaultOption[T any] struct {
	defaultValue T
}

func (o DefaultOption[T]) Apply(field interface{}) {
	// Try to use interface first
	if setter, ok := field.(DefaultValueSetter[T]); ok {
		setter.SetDefaultValue(o.defaultValue)
		return
	}

	// Fallback to reflection for types that don't implement the interface
	fieldValue := reflect.ValueOf(field)
	if fieldValue.Kind() == reflect.Ptr {
		fieldValue = fieldValue.Elem()
	}

	defaultValueField := fieldValue.FieldByName("defaultValue")
	hasDefaultField := fieldValue.FieldByName("hasDefault")

	if defaultValueField.IsValid() && defaultValueField.CanSet() {
		defaultValueField.Set(reflect.ValueOf(o.defaultValue))
	}
	if hasDefaultField.IsValid() && hasDefaultField.CanSet() {
		hasDefaultField.Set(reflect.ValueOf(true))
	}
}

// WithDefault creates a default value option
func WithDefault[T any](defaultValue T) Option {
	return DefaultOption[T]{defaultValue: defaultValue}
}
