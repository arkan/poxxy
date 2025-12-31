package poxxy

import (
	"fmt"
	"reflect"
)

// FieldBuilder provides a fluent API for building fields.
type FieldBuilder[T any] struct {
	fieldName      string
	dest           *T
	validators     []Validator[T]
	transformers   []Transformer[T]
	defaultValue   *T
	defaultOnEmpty bool
	strictType     bool
	desc           string
	absent         bool
	null           bool
}

// Field creates a new field builder for a value type.
func Field[T any](name string, dest *T) *FieldBuilder[T] {
	return &FieldBuilder[T]{
		fieldName: name,
		dest:      dest,
	}
}

// Validate adds validators to the field.
func (b *FieldBuilder[T]) Validate(validators ...Validator[T]) *FieldBuilder[T] {
	b.validators = append(b.validators, validators...)
	return b
}

// Transform adds transformers to the field.
func (b *FieldBuilder[T]) Transform(transformers ...Transformer[T]) *FieldBuilder[T] {
	b.transformers = append(b.transformers, transformers...)
	return b
}

// Default sets a default value (applied if field is absent).
func (b *FieldBuilder[T]) Default(value T) *FieldBuilder[T] {
	b.defaultValue = &value
	return b
}

// DefaultOnEmpty sets a default value (applied if field is absent OR empty).
func (b *FieldBuilder[T]) DefaultOnEmpty(value T) *FieldBuilder[T] {
	b.defaultValue = &value
	b.defaultOnEmpty = true
	return b
}

// StrictType disables automatic type conversion for this field.
func (b *FieldBuilder[T]) StrictType() *FieldBuilder[T] {
	b.strictType = true
	return b
}

// Describe sets a description for the field.
func (b *FieldBuilder[T]) Describe(desc string) *FieldBuilder[T] {
	b.desc = desc
	return b
}

// Implement field interface for FieldBuilder

func (b *FieldBuilder[T]) name() string {
	return b.fieldName
}

func (b *FieldBuilder[T]) description() string {
	return b.desc
}

func (b *FieldBuilder[T]) isAbsent() bool {
	return b.absent
}

func (b *FieldBuilder[T]) isNull() bool {
	return b.null
}

func (b *FieldBuilder[T]) assign(data map[string]any, path string, schema *Schema) error {
	fullPath := fieldPath(path, b.fieldName)

	rawValue, exists := data[b.fieldName]
	if !exists {
		b.absent = true
		if b.defaultValue != nil {
			*b.dest = *b.defaultValue
			schema.log("default applied", "field", fullPath, "value", *b.defaultValue)
		}
		return nil
	}

	if rawValue == nil {
		b.null = true
		if b.defaultValue != nil {
			*b.dest = *b.defaultValue
			schema.log("default applied (null)", "field", fullPath, "value", *b.defaultValue)
		}
		return nil
	}

	// Convert value
	value, err := convertValue[T](rawValue, b.strictType || schema.strictTypes)
	if err != nil {
		return &ValidationError{
			Field:   fullPath,
			Value:   rawValue,
			Rule:    "type",
			Message: fmt.Sprintf("%s: %v", fullPath, err),
			Cause:   err,
		}
	}

	// Check if empty and defaultOnEmpty is set
	if b.defaultOnEmpty && b.defaultValue != nil && isZero(value) {
		value = *b.defaultValue
		schema.log("default applied (empty)", "field", fullPath, "value", value)
	}

	// Apply transformers
	for _, t := range b.transformers {
		schema.log("transforming", "field", fullPath, "transformer", t.Name())
		value, err = t.Transform(value)
		if err != nil {
			return &ValidationError{
				Field:   fullPath,
				Value:   value,
				Rule:    "transform",
				Message: fmt.Sprintf("%s: transform error: %v", fullPath, err),
				Cause:   err,
			}
		}
	}

	*b.dest = value
	return nil
}

func (b *FieldBuilder[T]) validate(path string, schema *Schema) *ValidationErrors {
	fullPath := fieldPath(path, b.fieldName)
	errs := &ValidationErrors{}

	for _, v := range b.validators {
		schema.log("validating", "field", fullPath, "validator", v.Rule())
		if err := v.Validate(*b.dest, fullPath); err != nil {
			if ve, ok := err.(*ValidationError); ok {
				errs.Add(ve)
			} else {
				errs.Add(&ValidationError{
					Field:   fullPath,
					Value:   *b.dest,
					Rule:    v.Rule(),
					Message: err.Error(),
					Cause:   err,
				})
			}
		}
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// PointerBuilder provides a fluent API for building optional pointer fields.
type PointerBuilder[T any] struct {
	fieldName      string
	dest           **T
	validators     []Validator[T]
	transformers   []Transformer[T]
	defaultValue   *T
	defaultOnEmpty bool
	strictType     bool
	desc           string
	absent         bool
	null           bool
}

// Pointer creates a new field builder for an optional (nullable) type.
func Pointer[T any](name string, dest **T) *PointerBuilder[T] {
	return &PointerBuilder[T]{
		fieldName: name,
		dest:      dest,
	}
}

func (b *PointerBuilder[T]) Validate(validators ...Validator[T]) *PointerBuilder[T] {
	b.validators = append(b.validators, validators...)
	return b
}

func (b *PointerBuilder[T]) Transform(transformers ...Transformer[T]) *PointerBuilder[T] {
	b.transformers = append(b.transformers, transformers...)
	return b
}

func (b *PointerBuilder[T]) Default(value T) *PointerBuilder[T] {
	b.defaultValue = &value
	return b
}

func (b *PointerBuilder[T]) DefaultOnEmpty(value T) *PointerBuilder[T] {
	b.defaultValue = &value
	b.defaultOnEmpty = true
	return b
}

func (b *PointerBuilder[T]) StrictType() *PointerBuilder[T] {
	b.strictType = true
	return b
}

func (b *PointerBuilder[T]) Describe(desc string) *PointerBuilder[T] {
	b.desc = desc
	return b
}

// pointerField is the internal representation of a pointer field.
type pointerField[T any] struct {
	fieldName      string
	dest           **T
	validators     []Validator[T]
	transformers   []Transformer[T]
	defaultValue   *T
	defaultOnEmpty bool
	strictType     bool
	desc           string
	absent         bool
	null           bool
}

func (b *PointerBuilder[T]) name() string {
	return b.fieldName
}

func (b *PointerBuilder[T]) description() string {
	return b.desc
}

func (b *PointerBuilder[T]) isAbsent() bool {
	return b.absent
}

func (b *PointerBuilder[T]) isNull() bool {
	return b.null
}

func (b *PointerBuilder[T]) assign(data map[string]any, path string, schema *Schema) error {
	fullPath := fieldPath(path, b.fieldName)

	rawValue, exists := data[b.fieldName]
	if !exists {
		b.absent = true
		if b.defaultValue != nil {
			*b.dest = b.defaultValue
			schema.log("default applied", "field", fullPath, "value", *b.defaultValue)
		}
		return nil
	}

	if rawValue == nil {
		b.null = true
		if b.defaultValue != nil {
			*b.dest = b.defaultValue
			schema.log("default applied (null)", "field", fullPath, "value", *b.defaultValue)
		}
		return nil
	}

	value, err := convertValue[T](rawValue, b.strictType || schema.strictTypes)
	if err != nil {
		return &ValidationError{
			Field:   fullPath,
			Value:   rawValue,
			Rule:    "type",
			Message: fmt.Sprintf("%s: %v", fullPath, err),
			Cause:   err,
		}
	}

	if b.defaultOnEmpty && b.defaultValue != nil && isZero(value) {
		value = *b.defaultValue
		schema.log("default applied (empty)", "field", fullPath, "value", value)
	}

	for _, t := range b.transformers {
		schema.log("transforming", "field", fullPath, "transformer", t.Name())
		value, err = t.Transform(value)
		if err != nil {
			return &ValidationError{
				Field:   fullPath,
				Value:   value,
				Rule:    "transform",
				Message: fmt.Sprintf("%s: transform error: %v", fullPath, err),
				Cause:   err,
			}
		}
	}

	*b.dest = &value
	return nil
}

func (b *PointerBuilder[T]) validate(path string, schema *Schema) *ValidationErrors {
	fullPath := fieldPath(path, b.fieldName)

	// Skip validation if nil (optional field)
	if *b.dest == nil {
		return nil
	}

	errs := &ValidationErrors{}
	for _, v := range b.validators {
		schema.log("validating", "field", fullPath, "validator", v.Rule())
		if err := v.Validate(**b.dest, fullPath); err != nil {
			if ve, ok := err.(*ValidationError); ok {
				errs.Add(ve)
			} else {
				errs.Add(&ValidationError{
					Field:   fullPath,
					Value:   **b.dest,
					Rule:    v.Rule(),
					Message: err.Error(),
					Cause:   err,
				})
			}
		}
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// SliceBuilder provides a fluent API for building slice fields.
type SliceBuilder[T any] struct {
	fieldName    string
	dest         *[]T
	validators   []Validator[[]T]
	itemCallback func(*T, *Schema)
	strictType   bool
	desc         string
	absent       bool
	null         bool
}

// Slice creates a new field builder for a slice type.
func Slice[T any](name string, dest *[]T, itemCallback ...func(*T, *Schema)) *SliceBuilder[T] {
	b := &SliceBuilder[T]{
		fieldName: name,
		dest:      dest,
	}
	if len(itemCallback) > 0 {
		b.itemCallback = itemCallback[0]
	}
	return b
}

func (b *SliceBuilder[T]) Validate(validators ...Validator[[]T]) *SliceBuilder[T] {
	b.validators = append(b.validators, validators...)
	return b
}

func (b *SliceBuilder[T]) StrictType() *SliceBuilder[T] {
	b.strictType = true
	return b
}

func (b *SliceBuilder[T]) Describe(desc string) *SliceBuilder[T] {
	b.desc = desc
	return b
}

func (b *SliceBuilder[T]) name() string {
	return b.fieldName
}

func (b *SliceBuilder[T]) description() string {
	return b.desc
}

func (b *SliceBuilder[T]) isAbsent() bool {
	return b.absent
}

func (b *SliceBuilder[T]) isNull() bool {
	return b.null
}

func (b *SliceBuilder[T]) assign(data map[string]any, path string, schema *Schema) error {
	fullPath := fieldPath(path, b.fieldName)

	rawValue, exists := data[b.fieldName]
	if !exists {
		b.absent = true
		return nil
	}
	if rawValue == nil {
		b.null = true
		return nil
	}

	rawSlice, ok := rawValue.([]any)
	if !ok {
		return &ValidationError{
			Field:   fullPath,
			Value:   rawValue,
			Rule:    "type",
			Message: fmt.Sprintf("%s: expected array, got %T", fullPath, rawValue),
		}
	}

	result := make([]T, 0, len(rawSlice))

	for i, rawItem := range rawSlice {
		itemPath := indexPath(fullPath, i)

		// If we have a callback for complex types
		if b.itemCallback != nil {
			itemMap, ok := rawItem.(map[string]any)
			if !ok {
				return &ValidationError{
					Field:   itemPath,
					Value:   rawItem,
					Rule:    "type",
					Message: fmt.Sprintf("%s: expected object, got %T", itemPath, rawItem),
				}
			}

			var item T
			itemSchema := &Schema{
				fields:      make([]field, 0),
				fieldMap:    make(map[string]field),
				logger:      schema.logger,
				strictTypes: schema.strictTypes,
			}

			// Call the callback to set up the item schema
			b.itemCallback(&item, itemSchema)

			// Parse the item
			if err := itemSchema.ParseMap(itemMap); err != nil {
				// Adjust error paths
				if ve, ok := err.(*ValidationErrors); ok {
					for _, e := range ve.Errors {
						e.Field = itemPath + "." + e.Field
					}
					return ve
				}
				return err
			}

			result = append(result, item)
		} else {
			// Simple type conversion
			item, err := convertValue[T](rawItem, b.strictType || schema.strictTypes)
			if err != nil {
				return &ValidationError{
					Field:   itemPath,
					Value:   rawItem,
					Rule:    "type",
					Message: fmt.Sprintf("%s: %v", itemPath, err),
					Cause:   err,
				}
			}
			result = append(result, item)
		}
	}

	*b.dest = result
	return nil
}

func (b *SliceBuilder[T]) validate(path string, schema *Schema) *ValidationErrors {
	fullPath := fieldPath(path, b.fieldName)
	errs := &ValidationErrors{}

	for _, v := range b.validators {
		schema.log("validating", "field", fullPath, "validator", v.Rule())
		if err := v.Validate(*b.dest, fullPath); err != nil {
			if ve, ok := err.(*ValidationError); ok {
				errs.Add(ve)
			} else {
				errs.Add(&ValidationError{
					Field:   fullPath,
					Value:   *b.dest,
					Rule:    v.Rule(),
					Message: err.Error(),
					Cause:   err,
				})
			}
		}
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// MapBuilder provides a fluent API for building map fields.
type MapBuilder[K comparable, V any] struct {
	fieldName  string
	dest       *map[K]V
	validators []Validator[map[K]V]
	strictType bool
	desc       string
	absent     bool
	null       bool
}

// Map creates a new field builder for a map type.
func Map[K comparable, V any](name string, dest *map[K]V) *MapBuilder[K, V] {
	return &MapBuilder[K, V]{
		fieldName: name,
		dest:      dest,
	}
}

func (b *MapBuilder[K, V]) Validate(validators ...Validator[map[K]V]) *MapBuilder[K, V] {
	b.validators = append(b.validators, validators...)
	return b
}

func (b *MapBuilder[K, V]) StrictType() *MapBuilder[K, V] {
	b.strictType = true
	return b
}

func (b *MapBuilder[K, V]) Describe(desc string) *MapBuilder[K, V] {
	b.desc = desc
	return b
}

func (b *MapBuilder[K, V]) name() string {
	return b.fieldName
}

func (b *MapBuilder[K, V]) description() string {
	return b.desc
}

func (b *MapBuilder[K, V]) isAbsent() bool {
	return b.absent
}

func (b *MapBuilder[K, V]) isNull() bool {
	return b.null
}

func (b *MapBuilder[K, V]) assign(data map[string]any, path string, schema *Schema) error {
	fullPath := fieldPath(path, b.fieldName)

	rawValue, exists := data[b.fieldName]
	if !exists {
		b.absent = true
		return nil
	}
	if rawValue == nil {
		b.null = true
		return nil
	}

	rawMap, ok := rawValue.(map[string]any)
	if !ok {
		return &ValidationError{
			Field:   fullPath,
			Value:   rawValue,
			Rule:    "type",
			Message: fmt.Sprintf("%s: expected object, got %T", fullPath, rawValue),
		}
	}

	result := make(map[K]V)

	for rawKey, rawVal := range rawMap {
		key, err := convertValue[K](rawKey, b.strictType || schema.strictTypes)
		if err != nil {
			return &ValidationError{
				Field:   fullPath + "." + rawKey,
				Value:   rawKey,
				Rule:    "type",
				Message: fmt.Sprintf("%s: invalid key: %v", fullPath, err),
				Cause:   err,
			}
		}

		val, err := convertValue[V](rawVal, b.strictType || schema.strictTypes)
		if err != nil {
			return &ValidationError{
				Field:   fullPath + "." + rawKey,
				Value:   rawVal,
				Rule:    "type",
				Message: fmt.Sprintf("%s.%s: %v", fullPath, rawKey, err),
				Cause:   err,
			}
		}

		result[key] = val
	}

	*b.dest = result
	return nil
}

func (b *MapBuilder[K, V]) validate(path string, schema *Schema) *ValidationErrors {
	fullPath := fieldPath(path, b.fieldName)
	errs := &ValidationErrors{}

	for _, v := range b.validators {
		schema.log("validating", "field", fullPath, "validator", v.Rule())
		if err := v.Validate(*b.dest, fullPath); err != nil {
			if ve, ok := err.(*ValidationError); ok {
				errs.Add(ve)
			} else {
				errs.Add(&ValidationError{
					Field:   fullPath,
					Value:   *b.dest,
					Rule:    v.Rule(),
					Message: err.Error(),
					Cause:   err,
				})
			}
		}
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// structField represents a nested struct field.
type structField struct {
	fieldName string
	dest      any
	schema    *Schema
	desc      string
	absent    bool
	null      bool
}

// Struct creates a nested struct field with a callback for defining the sub-schema.
func Struct[T any](name string, dest *T, callback func(*Schema)) field {
	subSchema := &Schema{
		fields:   make([]field, 0),
		fieldMap: make(map[string]field),
	}
	callback(subSchema)

	return &structField{
		fieldName: name,
		dest:      dest,
		schema:    subSchema,
		desc:      "",
	}
}

func (f *structField) name() string {
	return f.fieldName
}

func (f *structField) description() string {
	return f.desc
}

func (f *structField) isAbsent() bool {
	return f.absent
}

func (f *structField) isNull() bool {
	return f.null
}

func (f *structField) assign(data map[string]any, path string, schema *Schema) error {
	fullPath := fieldPath(path, f.fieldName)

	rawValue, exists := data[f.fieldName]
	if !exists {
		f.absent = true
		return nil
	}

	if rawValue == nil {
		f.null = true
		return nil
	}

	rawMap, ok := rawValue.(map[string]any)
	if !ok {
		return &ValidationError{
			Field:   fullPath,
			Value:   rawValue,
			Rule:    "type",
			Message: fmt.Sprintf("%s: expected object, got %T", fullPath, rawValue),
		}
	}

	// Propagate logger and strictTypes
	f.schema.logger = schema.logger
	f.schema.strictTypes = schema.strictTypes

	// Assign nested fields
	for _, nested := range f.schema.fields {
		if err := nested.assign(rawMap, fullPath, f.schema); err != nil {
			return err
		}
	}

	return nil
}

func (f *structField) validate(path string, schema *Schema) *ValidationErrors {
	fullPath := fieldPath(path, f.fieldName)

	if f.absent || f.null {
		return nil
	}

	errs := &ValidationErrors{}

	for _, nested := range f.schema.fields {
		if fieldErrs := nested.validate(fullPath, f.schema); fieldErrs != nil {
			errs.Merge(fieldErrs)
		}
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// ConvertBuilder provides a fluent API for type conversion fields.
type ConvertBuilder[From, To any] struct {
	fieldName    string
	dest         *To
	converter    func(From) (To, error)
	validators   []Validator[To]
	transformers []Transformer[To]
	defaultValue *To
	strictType   bool
	desc         string
	absent       bool
	null         bool
}

// Convert creates a field that converts from one type to another.
func Convert[From, To any](name string, dest *To, converter func(From) (To, error)) *ConvertBuilder[From, To] {
	return &ConvertBuilder[From, To]{
		fieldName: name,
		dest:      dest,
		converter: converter,
	}
}

func (b *ConvertBuilder[From, To]) Validate(validators ...Validator[To]) *ConvertBuilder[From, To] {
	b.validators = append(b.validators, validators...)
	return b
}

func (b *ConvertBuilder[From, To]) Transform(transformers ...Transformer[To]) *ConvertBuilder[From, To] {
	b.transformers = append(b.transformers, transformers...)
	return b
}

func (b *ConvertBuilder[From, To]) Default(value To) *ConvertBuilder[From, To] {
	b.defaultValue = &value
	return b
}

func (b *ConvertBuilder[From, To]) StrictType() *ConvertBuilder[From, To] {
	b.strictType = true
	return b
}

func (b *ConvertBuilder[From, To]) Describe(desc string) *ConvertBuilder[From, To] {
	b.desc = desc
	return b
}

func (b *ConvertBuilder[From, To]) name() string {
	return b.fieldName
}

func (b *ConvertBuilder[From, To]) description() string {
	return b.desc
}

func (b *ConvertBuilder[From, To]) isAbsent() bool {
	return b.absent
}

func (b *ConvertBuilder[From, To]) isNull() bool {
	return b.null
}

func (b *ConvertBuilder[From, To]) assign(data map[string]any, path string, schema *Schema) error {
	fullPath := fieldPath(path, b.fieldName)

	rawValue, exists := data[b.fieldName]
	if !exists {
		b.absent = true
		if b.defaultValue != nil {
			*b.dest = *b.defaultValue
			schema.log("default applied", "field", fullPath, "value", *b.defaultValue)
		}
		return nil
	}
	if rawValue == nil {
		b.null = true
		if b.defaultValue != nil {
			*b.dest = *b.defaultValue
			schema.log("default applied (null)", "field", fullPath, "value", *b.defaultValue)
		}
		return nil
	}

	// First convert to From type
	fromValue, err := convertValue[From](rawValue, b.strictType || schema.strictTypes)
	if err != nil {
		return &ValidationError{
			Field:   fullPath,
			Value:   rawValue,
			Rule:    "type",
			Message: fmt.Sprintf("%s: %v", fullPath, err),
			Cause:   err,
		}
	}

	// Then apply custom converter
	toValue, err := b.converter(fromValue)
	if err != nil {
		return &ValidationError{
			Field:   fullPath,
			Value:   fromValue,
			Rule:    "convert",
			Message: fmt.Sprintf("%s: conversion error: %v", fullPath, err),
			Cause:   err,
		}
	}

	// Apply transformers
	for _, t := range b.transformers {
		schema.log("transforming", "field", fullPath, "transformer", t.Name())
		toValue, err = t.Transform(toValue)
		if err != nil {
			return &ValidationError{
				Field:   fullPath,
				Value:   toValue,
				Rule:    "transform",
				Message: fmt.Sprintf("%s: transform error: %v", fullPath, err),
				Cause:   err,
			}
		}
	}

	*b.dest = toValue
	return nil
}

func (b *ConvertBuilder[From, To]) validate(path string, schema *Schema) *ValidationErrors {
	fullPath := fieldPath(path, b.fieldName)
	errs := &ValidationErrors{}

	for _, v := range b.validators {
		schema.log("validating", "field", fullPath, "validator", v.Rule())
		if err := v.Validate(*b.dest, fullPath); err != nil {
			if ve, ok := err.(*ValidationError); ok {
				errs.Add(ve)
			} else {
				errs.Add(&ValidationError{
					Field:   fullPath,
					Value:   *b.dest,
					Rule:    v.Rule(),
					Message: err.Error(),
					Cause:   err,
				})
			}
		}
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// isZero checks if a value is its zero value.
func isZero[T any](v T) bool {
	return reflect.ValueOf(&v).Elem().IsZero()
}
