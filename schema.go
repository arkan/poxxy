package poxxy

import (
	"encoding/json"
	"fmt"
	"io"
	"net/http"
)

// MaxBodySize is the maximum size of the body of an HTTP request
// You can change this value to limit the size of the body of an HTTP request
const MaxBodySize = 5 << 20 // 5MB limit

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
// It supports application/json and application/x-www-form-urlencoded
// It will return an error if the content type is not supported
func (s *Schema) ApplyHTTPRequest(r *http.Request) error {
	contentType := r.Header.Get("Content-Type")
	switch contentType {
	case "application/json":
		body, err := io.ReadAll(io.LimitReader(r.Body, MaxBodySize))
		if err != nil {
			return fmt.Errorf("failed to read request body: %w", err)
		}

		var data map[string]interface{}
		if err := json.Unmarshal(body, &data); err != nil {
			return fmt.Errorf("failed to unmarshal request body: %w", err)
		}
		return s.Apply(data)
	case "application/x-www-form-urlencoded":
		if err := r.ParseForm(); err != nil {
			return fmt.Errorf("failed to parse form: %w", err)
		}

		form := make(map[string]interface{})

		// Note: we are using Postform and not Form because we don't want to include
		// the data from the url query params.
		// See: https://pkg.go.dev/net/http#Request.PostForm
		for key, values := range r.PostForm {
			// We only support the first value of each form field
			form[key] = values[0]
		}

		return s.Apply(form)
	default:
		// We parse request url query params
		params := make(map[string]interface{})
		for key, values := range r.URL.Query() {
			params[key] = values[0]
		}
		return s.Apply(params)
	}
}

// ApplyJSON assigns data from a JSON string to a schema
func (s *Schema) ApplyJSON(jsonData []byte) error {
	var data map[string]interface{}

	if err := json.Unmarshal(jsonData, &data); err != nil {
		return fmt.Errorf("failed to unmarshal request body: %w", err)
	}

	return s.Apply(data)
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
			errors = append(errors, FieldError{Field: field.Name(), Error: err, Description: field.Description()})
		}
	}

	// If there are any errors, return them
	if len(errors) > 0 {
		return errors
	}

	// Second pass: validate
	for _, field := range s.fields {
		if err := field.Validate(s); err != nil {
			errors = append(errors, FieldError{Field: field.Name(), Error: err, Description: field.Description()})
		}
	}

	// If there are any errors, return them
	if len(errors) > 0 {
		return errors
	}

	return nil
}

// IsFieldPresent checks if a field was present in the input data
func (s *Schema) IsFieldPresent(fieldName string) bool {
	return s.presentFields[fieldName]
}

func (s *Schema) SetFieldPresent(fieldName string) {
	s.presentFields[fieldName] = true
}

func WithSchema(schema *Schema, field Field) {
	schema.fields = append(schema.fields, field)
}

type SubSchemaOption[T any] struct {
	callback func(*Schema, *T)
}

type SubSchemaInterface[T any] interface {
	SetCallback(func(*Schema, *T))
}

type SubSchemaMapInterface[K comparable, V any] interface {
	SetCallback(func(*Schema, K, V))
}

func (o SubSchemaOption[T]) Apply(field interface{}) {
	if f, ok := field.(SubSchemaInterface[T]); ok {
		f.SetCallback(o.callback)
	} else {
		panic(fmt.Sprintf("WithSubSchema doesn't support %T", field))
	}
}

func WithSubSchema[T any](callback func(*Schema, *T)) Option {
	return SubSchemaOption[T]{callback: callback}
}

type SubSchemaMapOption[K comparable, V any] struct {
	callback func(*Schema, K, V)
}

func (o SubSchemaMapOption[K, V]) Apply(field interface{}) {
	if f, ok := field.(SubSchemaMapInterface[K, V]); ok {
		f.SetCallback(o.callback)
	} else {
		panic(fmt.Sprintf("WithSubSchemaMap doesn't support %T", field))
	}
}

func WithSubSchemaMap[K comparable, V any](callback func(*Schema, K, V)) Option {
	return SubSchemaMapOption[K, V]{callback: callback}
}
