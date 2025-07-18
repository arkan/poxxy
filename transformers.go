package poxxy

import (
	"strings"
	"unicode"
)

// Transformer represents a function that transforms a value before assignment and validation
type Transformer[T any] interface {
	Transform(value T) (T, error)
}

// TransformerFn is a function that implements Transformer
type TransformerFn[T any] struct {
	fn func(T) (T, error)
}

func (t TransformerFn[T]) Transform(value T) (T, error) {
	return t.fn(value)
}

// TransformerOption holds transformers
type TransformerOption[T any] struct {
	transformers []Transformer[T]
}

func (o TransformerOption[T]) Apply(field interface{}) {
	// Try to use type assertion first
	if valueField, ok := field.(*ValueField[T]); ok {
		for _, transformer := range o.transformers {
			valueField.AddTransformer(transformer)
		}
		return
	}
	if pointerField, ok := field.(*PointerField[T]); ok {
		for _, transformer := range o.transformers {
			pointerField.AddTransformer(transformer)
		}
		return
	}
}

// WithTransformers creates a transformers option
func WithTransformers[T any](transformers ...Transformer[T]) Option {
	return TransformerOption[T]{transformers: transformers}
}

// Built-in transformers

// ToUpper transforms a string to uppercase
func ToUpper() Transformer[string] {
	return TransformerFn[string]{
		fn: func(value string) (string, error) {
			return strings.ToUpper(value), nil
		},
	}
}

// ToLower transforms a string to lowercase
func ToLower() Transformer[string] {
	return TransformerFn[string]{
		fn: func(value string) (string, error) {
			return strings.ToLower(value), nil
		},
	}
}

// TrimSpace removes leading and trailing whitespace
func TrimSpace() Transformer[string] {
	return TransformerFn[string]{
		fn: func(value string) (string, error) {
			return strings.TrimSpace(value), nil
		},
	}
}

// TitleCase transforms a string to title case
func TitleCase() Transformer[string] {
	return TransformerFn[string]{
		fn: func(value string) (string, error) {
			return strings.Title(strings.ToLower(value)), nil
		},
	}
}

// Capitalize transforms a string to capitalize first letter
func Capitalize() Transformer[string] {
	return TransformerFn[string]{
		fn: func(value string) (string, error) {
			if len(value) == 0 {
				return value, nil
			}
			return string(unicode.ToUpper(rune(value[0]))) + strings.ToLower(value[1:]), nil
		},
	}
}

// SanitizeEmail normalizes email addresses
func SanitizeEmail() Transformer[string] {
	return TransformerFn[string]{
		fn: func(value string) (string, error) {
			return strings.ToLower(strings.TrimSpace(value)), nil
		},
	}
}

// Custom transformer creator
func CustomTransformer[T any](transform func(T) (T, error)) Transformer[T] {
	return TransformerFn[T]{fn: transform}
}
