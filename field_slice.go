package poxxy

import (
	"fmt"
	"reflect"
)

// SliceField represents a slice field where each element is a struct
type SliceField[T any] struct {
	name         string
	description  string
	ptr          *[]T
	callback     func(*Schema, *T)
	Validators   []Validator
	wasAssigned  bool // Track if a non-nil value was assigned
	defaultValue []T
	hasDefault   bool
	transformers []Transformer[[]T]
}

func (f *SliceField[T]) Name() string {
	return f.name
}

func (f *SliceField[T]) Value() interface{} {
	if f.ptr == nil {
		return nil
	}
	if !f.wasAssigned {
		return nil
	}
	return *f.ptr
}

func (f *SliceField[T]) Description() string {
	return f.description
}

func (f *SliceField[T]) SetDescription(description string) {
	f.description = description
}

func (f *SliceField[T]) AddTransformer(transformer Transformer[[]T]) {
	f.transformers = append(f.transformers, transformer)
}

func (f *SliceField[T]) SetDefaultValue(defaultValue []T) {
	f.defaultValue = defaultValue
	f.hasDefault = true
}

func (f *SliceField[T]) Assign(data map[string]interface{}, schema *Schema) error {
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

	// Accept []interface{}, []map[string]interface{}, or any slice/array via reflection
	var slice []interface{}

	switch v := value.(type) {
	case []interface{}:
		slice = v
	case []map[string]interface{}:
		slice = make([]interface{}, len(v))
		for i, item := range v {
			slice[i] = item
		}
	default:
		rValue := reflect.ValueOf(value)
		if rValue.Kind() != reflect.Slice && rValue.Kind() != reflect.Array {
			return fmt.Errorf("expected slice, got %T", value)
		}
		slice = make([]interface{}, rValue.Len())
		for i := 0; i < rValue.Len(); i++ {
			slice[i] = rValue.Index(i).Interface()
		}
	}

	result := make([]T, len(slice))

	for i, item := range slice {
		switch v := item.(type) {
		case map[string]interface{}:
			var element T
			subSchema := NewSchema()
			if f.callback != nil {
				f.callback(subSchema, &element)
			}
			if err := subSchema.Apply(v); err != nil {
				return fmt.Errorf("element %d: %v", i, err)
			}
			result[i] = element
		default:
			converted, err := convertValue[T](v)
			if err != nil {
				return fmt.Errorf("element %d: %v", i, err)
			}
			result[i] = converted
		}
	}

	// Apply transformers
	transformed := result
	for _, transformer := range f.transformers {
		var err error
		transformed, err = transformer.Transform(transformed)
		if err != nil {
			return fmt.Errorf("transformer failed: %v", err)
		}
		result = transformed
	}

	*f.ptr = result
	f.wasAssigned = true
	return nil
}

func (f *SliceField[T]) Validate(schema *Schema) error {
	return validateFieldValidators(f.Validators, *f.ptr, f.name, schema)
}

func (f *SliceField[T]) SetCallback(callback func(*Schema, *T)) {
	f.callback = callback
}

// Slice creates a slice field.
func Slice[T any](name string, ptr *[]T, opts ...Option) Field {
	field := &SliceField[T]{
		name: name,
		ptr:  ptr,
	}

	for _, opt := range opts {
		opt.Apply(field)
	}

	return field
}
