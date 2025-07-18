package poxxy

import (
	"fmt"
	"reflect"
)

// ArrayField represents an array field
type ArrayField[T any] struct {
	name         string
	description  string
	ptr          interface{} // *[N]T
	Validators   []Validator
	wasAssigned  bool        // Track if a non-nil value was assigned
	defaultValue interface{} // [N]T
	hasDefault   bool
	transformers []Transformer[interface{}]
}

func (f *ArrayField[T]) Name() string {
	return f.name
}

func (f *ArrayField[T]) Value() interface{} {
	if f.ptr == nil {
		return nil
	}
	if !f.wasAssigned {
		return nil
	}
	return f.ptr
}

func (f *ArrayField[T]) Description() string {
	return f.description
}

func (f *ArrayField[T]) SetDescription(description string) {
	f.description = description
}

func (f *ArrayField[T]) AddTransformer(transformer Transformer[interface{}]) {
	f.transformers = append(f.transformers, transformer)
}

func (f *ArrayField[T]) SetDefaultValue(defaultValue interface{}) {
	f.defaultValue = defaultValue
	f.hasDefault = true
}

func (f *ArrayField[T]) Assign(data map[string]interface{}, schema *Schema) error {
	value, exists := data[f.name]
	if !exists {
		// Apply default value if available
		if f.hasDefault {
			ptrValue := reflect.ValueOf(f.ptr)
			defaultValue := reflect.ValueOf(f.defaultValue)
			ptrValue.Elem().Set(defaultValue)
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

	// Get the array pointer and its element type
	ptrValue := reflect.ValueOf(f.ptr)
	if ptrValue.Kind() != reflect.Ptr {
		return fmt.Errorf("expected pointer to array")
	}

	arrayValue := ptrValue.Elem()
	if arrayValue.Kind() != reflect.Array {
		return fmt.Errorf("expected array type")
	}

	// Convert source value to slice for easier handling
	sourceValue := reflect.ValueOf(value)
	if sourceValue.Kind() != reflect.Slice && sourceValue.Kind() != reflect.Array {
		return fmt.Errorf("source value must be slice or array")
	}

	// Check length
	if sourceValue.Len() != arrayValue.Len() {
		return fmt.Errorf("array length mismatch: expected %d, got %d", arrayValue.Len(), sourceValue.Len())
	}

	// Copy elements
	for i := 0; i < sourceValue.Len(); i++ {
		srcElem := sourceValue.Index(i).Interface()
		converted, err := convertValue[T](srcElem)
		if err != nil {
			return fmt.Errorf("element %d: %v", i, err)
		}
		arrayValue.Index(i).Set(reflect.ValueOf(converted))
	}

	// Apply transformers
	transformed := arrayValue.Interface()
	for _, transformer := range f.transformers {
		var err error
		transformed, err = transformer.Transform(transformed)
		if err != nil {
			return fmt.Errorf("transformer failed: %v", err)
		}
		// Update the array with transformed value
		reflect.ValueOf(f.ptr).Elem().Set(reflect.ValueOf(transformed))
	}

	f.wasAssigned = true
	return nil
}

func (f *ArrayField[T]) Validate(schema *Schema) error {
	arrayValue := reflect.ValueOf(f.ptr).Elem()
	arrayInterface := arrayValue.Interface()

	return validateFieldValidators(f.Validators, arrayInterface, f.name, schema)
}

// AppendValidators implements ValidatorsAppender interface
func (f *ArrayField[T]) AppendValidators(validators []Validator) {
	f.Validators = append(f.Validators, validators...)
}

// Array creates an array field
func Array[T any](name string, ptr interface{}, opts ...Option) Field {
	field := &ArrayField[T]{
		name: name,
		ptr:  ptr,
	}

	for _, opt := range opts {
		opt.Apply(field)
	}

	return field
}
