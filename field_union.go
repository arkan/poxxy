package poxxy

import (
	"fmt"
	"reflect"
)

// UnionField represents a union/polymorphic field
type UnionField struct {
	name        string
	description string
	ptr         interface{}
	resolver    func(map[string]interface{}) (interface{}, error)
	wasAssigned bool // Track if a non-nil value was assigned
}

// Name returns the field name
func (f *UnionField) Name() string {
	return f.name
}

// Description returns the field description
func (f *UnionField) Description() string {
	return f.description
}

// SetDescription sets the field description
func (f *UnionField) SetDescription(description string) {
	f.description = description
}

// Value returns the current value of the field
func (f *UnionField) Value() interface{} {
	if f.ptr == nil {
		return nil
	}

	if !f.wasAssigned {
		return nil
	}

	return f.ptr
}

// Assign assigns a value to the field from the input data
func (f *UnionField) Assign(data map[string]interface{}, schema *Schema) error {
	value, exists := data[f.name]
	if !exists {
		return nil
	}

	schema.SetFieldPresent(f.name)

	if value == nil {
		f.wasAssigned = false
		return nil
	}

	mapData, ok := value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected object for union field")
	}

	// Use resolver to determine and create the correct type
	result, err := f.resolver(mapData)
	if err != nil {
		return err
	}

	// Assign the result to the pointer
	ptrValue := reflect.ValueOf(f.ptr)
	if ptrValue.Kind() != reflect.Ptr || ptrValue.Elem().Kind() != reflect.Interface {
		return fmt.Errorf("union field pointer must be pointer to interface")
	}

	ptrValue.Elem().Set(reflect.ValueOf(result))
	f.wasAssigned = true

	return nil
}

// Validate validates the field value using all registered validators
func (f *UnionField) Validate(schema *Schema) error {
	// Validation happens during assignment
	return nil
}

// Union creates a union field
func Union(name string, ptr interface{}, resolver func(map[string]interface{}) (interface{}, error)) Field {
	return &UnionField{
		name:     name,
		ptr:      ptr,
		resolver: resolver,
	}
}
