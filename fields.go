package poxxy

import (
	"fmt"
	"reflect"
)

// ArrayField represents an array field
type ArrayField[T any] struct {
	name       string
	ptr        interface{} // *[N]T
	Validators []Validator
}

func (f *ArrayField[T]) Name() string {
	return f.name
}

func (f *ArrayField[T]) Assign(data map[string]interface{}, schema *Schema) error {
	value, exists := data[f.name]
	if !exists {
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

	return nil
}

func (f *ArrayField[T]) Validate(schema *Schema) error {
	arrayValue := reflect.ValueOf(f.ptr).Elem()
	arrayInterface := arrayValue.Interface()

	for _, validator := range f.Validators {
		if err := validator.Validate(arrayInterface, f.name); err != nil {
			return err
		}
	}
	return nil
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

// SliceField represents a slice field
type SliceField[T any] struct {
	name       string
	ptr        *[]T
	Validators []Validator
}

func (f *SliceField[T]) Name() string {
	return f.name
}

func (f *SliceField[T]) Assign(data map[string]interface{}, schema *Schema) error {
	value, exists := data[f.name]
	if !exists {
		return nil
	}

	sourceValue := reflect.ValueOf(value)
	if sourceValue.Kind() != reflect.Slice && sourceValue.Kind() != reflect.Array {
		return fmt.Errorf("source value must be slice or array")
	}

	// Create new slice
	result := make([]T, sourceValue.Len())

	// Convert elements
	for i := 0; i < sourceValue.Len(); i++ {
		srcElem := sourceValue.Index(i).Interface()
		converted, err := convertValue[T](srcElem)
		if err != nil {
			return fmt.Errorf("element %d: %v", i, err)
		}
		result[i] = converted
	}

	*f.ptr = result
	return nil
}

func (f *SliceField[T]) Validate(schema *Schema) error {
	for _, validator := range f.Validators {
		if err := validator.Validate(*f.ptr, f.name); err != nil {
			return err
		}
	}
	return nil
}

// Slice creates a slice field
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

// UnionField represents a union/polymorphic field
type UnionField struct {
	name     string
	ptr      interface{}
	resolver func(map[string]interface{}) (interface{}, error)
}

func (f *UnionField) Name() string {
	return f.name
}

func (f *UnionField) Assign(data map[string]interface{}, schema *Schema) error {
	value, exists := data[f.name]
	if !exists {
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
	return nil
}

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

// StructField represents a struct field with callback
type StructField[T any] struct {
	name     string
	ptr      *T
	callback func(*Schema, *T)
}

func (f *StructField[T]) Name() string {
	return f.name
}

func (f *StructField[T]) Assign(data map[string]interface{}, schema *Schema) error {
	value, exists := data[f.name]
	if !exists {
		return nil
	}

	structData, ok := value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected object for struct field")
	}

	// Create a sub-schema and let the callback define it
	subSchema := NewSchema()
	f.callback(subSchema, f.ptr)

	// Assign and validate the struct data
	return subSchema.Apply(structData)
}

func (f *StructField[T]) Validate(schema *Schema) error {
	// Validation is done during assignment phase
	return nil
}

// Struct creates a struct field
func Struct[T any](name string, ptr *T, callback func(*Schema, *T)) Field {
	return &StructField[T]{
		name:     name,
		ptr:      ptr,
		callback: callback,
	}
}

// PointerField represents a pointer field
type PointerField[T any] struct {
	name       string
	ptr        **T
	Validators []Validator
	callback   func(*Schema, *T)
}

func (f *PointerField[T]) Name() string {
	return f.name
}

func (f *PointerField[T]) Assign(data map[string]interface{}, schema *Schema) error {
	value, exists := data[f.name]
	if !exists {
		// Optional field - leave as nil
		return nil
	}

	// Allocate new instance
	instance := new(T)
	*f.ptr = instance

	if f.callback != nil {
		// Handle struct pointer with callback
		structData, ok := value.(map[string]interface{})
		if !ok {
			return fmt.Errorf("expected object for struct pointer field")
		}

		subSchema := NewSchema()
		f.callback(subSchema, instance)
		return subSchema.Apply(structData)
	} else {
		// Handle simple pointer
		converted, err := convertValue[T](value)
		if err != nil {
			return fmt.Errorf("pointer field conversion failed: %v", err)
		}
		**f.ptr = converted
	}

	return nil
}

func (f *PointerField[T]) Validate(schema *Schema) error {
	if *f.ptr == nil {
		// Pointer is nil - skip validation unless required
		for _, validator := range f.Validators {
			if err := validator.Validate(nil, f.name); err != nil {
				return err
			}
		}
		return nil
	}

	// Validate the pointed value
	for _, validator := range f.Validators {
		if err := validator.Validate(**f.ptr, f.name); err != nil {
			return err
		}
	}
	return nil
}

// Pointer creates a pointer field
func Pointer[T any](name string, ptr **T, opts ...interface{}) Field {
	var validators []Validator
	var callback func(*Schema, *T)

	for _, opt := range opts {
		switch o := opt.(type) {
		case Option:
			if validatorOpt, ok := o.(ValidatorsOption); ok {
				validators = append(validators, validatorOpt.validators...)
			}
		case func(*Schema, *T):
			callback = o
		}
	}

	field := &PointerField[T]{
		name:       name,
		ptr:        ptr,
		Validators: validators,
		callback:   callback,
	}

	return field
}

// NestedMapField represents a nested map field
type NestedMapField[K comparable, V any] struct {
	name     string
	ptr      *map[K]V
	callback func(*Schema, K, *V)
}

func (f *NestedMapField[K, V]) Name() string {
	return f.name
}

func (f *NestedMapField[K, V]) Assign(data map[string]interface{}, schema *Schema) error {
	value, exists := data[f.name]
	if !exists {
		return nil
	}

	mapData, ok := value.(map[string]interface{})
	if !ok {
		return fmt.Errorf("expected map for nested map field")
	}

	result := make(map[K]V)

	for key, val := range mapData {
		// Convert key to type K
		convertedKey, err := convertValue[K](key)
		if err != nil {
			return fmt.Errorf("key conversion failed: %v", err)
		}

		// Convert value to type V
		convertedVal, err := convertValue[V](val)
		if err != nil {
			return fmt.Errorf("value conversion failed: %v", err)
		}

		result[convertedKey] = convertedVal

		// Run callback for validation if provided
		if f.callback != nil {
			subSchema := NewSchema()
			valCopy := convertedVal
			f.callback(subSchema, convertedKey, &valCopy)
		}
	}

	*f.ptr = result
	return nil
}

func (f *NestedMapField[K, V]) Validate(schema *Schema) error {
	// Validation happens during assignment
	return nil
}

// NestedMap creates a nested map field
func NestedMap[K comparable, V any](name string, ptr *map[K]V, callback func(*Schema, K, *V)) Field {
	return &NestedMapField[K, V]{
		name:     name,
		ptr:      ptr,
		callback: callback,
	}
}

// ValueFromField represents a field that validates a direct value
type ValueFromField[T any] struct {
	name       string
	value      interface{}
	Validators []Validator
}

func (f *ValueFromField[T]) Name() string {
	return f.name
}

func (f *ValueFromField[T]) Assign(data map[string]interface{}, schema *Schema) error {
	// ValueFrom doesn't assign, it validates existing values
	return nil
}

func (f *ValueFromField[T]) Validate(schema *Schema) error {
	converted, err := convertValue[T](f.value)
	if err != nil {
		return fmt.Errorf("type conversion failed: %v", err)
	}

	for _, validator := range f.Validators {
		if err := validator.Validate(converted, f.name); err != nil {
			return err
		}
	}
	return nil
}

// ValueFrom validates a direct value (used in nested map validation)
func ValueFrom[T any](name string, value interface{}, opts ...Option) Field {
	field := &ValueFromField[T]{
		name:  name,
		value: value,
	}

	for _, opt := range opts {
		opt.Apply(field)
	}

	return field
}

// TransformField represents a field with type transformation
type TransformField[From, To any] struct {
	name       string
	ptr        *To
	transform  func(From) (To, error)
	Validators []Validator
}

func (f *TransformField[From, To]) Name() string {
	return f.name
}

func (f *TransformField[From, To]) Assign(data map[string]interface{}, schema *Schema) error {
	value, exists := data[f.name]
	if !exists {
		return nil
	}

	// Convert to From type first
	fromValue, err := convertValue[From](value)
	if err != nil {
		return fmt.Errorf("transform source conversion failed: %v", err)
	}

	// Apply transformation
	toValue, err := f.transform(fromValue)
	if err != nil {
		return fmt.Errorf("transformation failed: %v", err)
	}

	*f.ptr = toValue
	return nil
}

func (f *TransformField[From, To]) Validate(schema *Schema) error {
	for _, validator := range f.Validators {
		if err := validator.Validate(*f.ptr, f.name); err != nil {
			return err
		}
	}
	return nil
}

// Transform creates a transformation field
func Transform[From, To any](name string, ptr *To, transform func(From) (To, error), opts ...Option) Field {
	field := &TransformField[From, To]{
		name:      name,
		ptr:       ptr,
		transform: transform,
	}

	for _, opt := range opts {
		opt.Apply(field)
	}

	return field
}

// SliceOfField represents a slice field where each element is a struct
type SliceOfField[T any] struct {
	name       string
	ptr        *[]T
	callback   func(*Schema, *T)
	Validators []Validator
}

func (f *SliceOfField[T]) Name() string {
	return f.name
}

func (f *SliceOfField[T]) Assign(data map[string]interface{}, schema *Schema) error {
	value, exists := data[f.name]
	if !exists {
		return nil
	}

	// Convert to slice of interface{} - handle different slice types
	var slice []interface{}

	switch v := value.(type) {
	case []interface{}:
		slice = v
	case []map[string]interface{}:
		// Convert []map[string]interface{} to []interface{}
		slice = make([]interface{}, len(v))
		for i, item := range v {
			slice[i] = item
		}
	default:
		// Try to use reflection to handle other slice types
		rValue := reflect.ValueOf(value)
		if rValue.Kind() != reflect.Slice {
			return fmt.Errorf("expected slice, got %T", value)
		}

		slice = make([]interface{}, rValue.Len())
		for i := 0; i < rValue.Len(); i++ {
			slice[i] = rValue.Index(i).Interface()
		}
	}

	// Create result slice
	result := make([]T, len(slice))

	// Process each element
	for i, item := range slice {
		itemMap, ok := item.(map[string]interface{})
		if !ok {
			return fmt.Errorf("element %d: expected map, got %T", i, item)
		}

		// Create a new instance for this element
		var element T

		// Create a sub-schema for this element
		subSchema := NewSchema()

		// Apply the callback to define the schema for this element
		if f.callback != nil {
			f.callback(subSchema, &element)
		}

		// Assign and validate this element
		if err := subSchema.Apply(itemMap); err != nil {
			return fmt.Errorf("element %d: %v", i, err)
		}

		result[i] = element
	}

	*f.ptr = result
	return nil
}

func (f *SliceOfField[T]) Validate(schema *Schema) error {
	return validateFieldValidators(f.Validators, *f.ptr, f.name, schema)
}

// SliceOf creates a slice field for structs with element-wise schema definition
func SliceOf[T any](name string, ptr *[]T, callback func(*Schema, *T), opts ...Option) Field {
	field := &SliceOfField[T]{
		name:     name,
		ptr:      ptr,
		callback: callback,
	}

	for _, opt := range opts {
		opt.Apply(field)
	}

	return field
}
