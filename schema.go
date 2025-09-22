package poxxy

import (
	"encoding/json"
	"fmt"
	"net/http"
)

// MaxBodySize is the maximum size of the body of an HTTP request
// You can change this value to limit the size of the body of an HTTP request
var MaxBodySize int64 = 5 << 20 // 5MB limit

// Schema represents a validation schema
type Schema struct {
	fields         []Field
	data           map[string]interface{}
	presentFields  map[string]bool // Track which fields were present in input data
	skipValidators bool
}

// NewSchema creates a new schema with the given fields
func NewSchema(fields ...Field) *Schema {
	return &Schema{
		fields:        fields,
		presentFields: make(map[string]bool),
	}
}

// SchemaOption represents a configuration option for a schema
type SchemaOption func(*Schema)

// WithSkipValidators creates a schema option to skip validation
func WithSkipValidators(skipValidators bool) SchemaOption {
	return func(s *Schema) {
		s.skipValidators = skipValidators
	}
}

type ContentTypeParsing uint8

const (
	_                                         = iota
	ContentTypeParsingAuto ContentTypeParsing = iota
	ContentTypeParsingJSON
	ContentTypeParsingForm
	ContentTypeParsingQuery
)

type HTTPRequestOption struct {
	MaxRequestBodySize int64
	ContentTypeParsing ContentTypeParsing
}

// ApplyHTTPRequest assigns data from an HTTP request to a schema
// It supports application/json and application/x-www-form-urlencoded
// It will return an error if the content type is not supported
func (s *Schema) ApplyHTTPRequest(w http.ResponseWriter, r *http.Request, httpRequestOption *HTTPRequestOption, options ...SchemaOption) error {
	if httpRequestOption == nil {
		httpRequestOption = &HTTPRequestOption{
			MaxRequestBodySize: MaxBodySize,
			ContentTypeParsing: ContentTypeParsingAuto,
		}
	}

	// Determine the content type parsing strategy depending on the content type header.
	// We only do this for ContentTypeParsingAuto.
	if httpRequestOption.ContentTypeParsing == ContentTypeParsingAuto {
		switch r.Header.Get("Content-Type") {
		case "application/json":
			httpRequestOption.ContentTypeParsing = ContentTypeParsingJSON
		case "application/x-www-form-urlencoded":
			httpRequestOption.ContentTypeParsing = ContentTypeParsingForm
		default:
			httpRequestOption.ContentTypeParsing = ContentTypeParsingQuery
		}
	}

	// Apply the content type parsing strategy.
	switch httpRequestOption.ContentTypeParsing {
	case ContentTypeParsingForm:
		if httpRequestOption.MaxRequestBodySize > 0 {
			// Limit the request body size
			r.Body = http.MaxBytesReader(w, r.Body, httpRequestOption.MaxRequestBodySize)
		}

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

		return s.Apply(form, options...)
	case ContentTypeParsingJSON:
		if httpRequestOption.MaxRequestBodySize > 0 {
			// Limit the request body size
			r.Body = http.MaxBytesReader(w, r.Body, httpRequestOption.MaxRequestBodySize)
		}

		var data map[string]interface{}
		if err := json.NewDecoder(r.Body).Decode(&data); err != nil {
			return fmt.Errorf("failed to unmarshal request body: %w", err)
		}

		return s.Apply(data, options...)
	default:
		// If the content type parsing strategy is not set, we fall through to the default case ContentTypeParsingQuery.
		fallthrough
	case ContentTypeParsingQuery:
		params := make(map[string]interface{})
		for key, values := range r.URL.Query() {
			params[key] = values[0]
		}

		return s.Apply(params, options...)
	}
}

// ApplyJSON assigns data from a JSON string to a schema
func (s *Schema) ApplyJSON(jsonData []byte, options ...SchemaOption) error {
	var data map[string]interface{}

	if err := json.Unmarshal(jsonData, &data); err != nil {
		return fmt.Errorf("failed to unmarshal request body: %w", err)
	}

	return s.Apply(data, options...)
}

// Apply assigns data to variables and validates them
func (s *Schema) Apply(data map[string]interface{}, options ...SchemaOption) error {
	s.data = data
	s.presentFields = make(map[string]bool)

	// Apply options to the schema
	for _, option := range options {
		option(s)
	}

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

	// If we skip validators, return any assignment errors
	if s.skipValidators {
		if len(errors) > 0 {
			return errors
		}
		return nil
	}

	// Second pass: validate (even if there were assignment errors)
	for _, field := range s.fields {
		if err := field.Validate(s); err != nil {
			errors = append(errors, FieldError{Field: field.Name(), Error: err, Description: field.Description()})
		}
	}

	// Return all errors (assignment + validation)
	if len(errors) > 0 {
		return errors
	}

	return nil
}

// GetFieldValue returns the value of a field by name
func (s *Schema) GetFieldValue(fieldName string) (interface{}, bool) {
	for _, field := range s.fields {
		if f, ok := field.(Field); ok && f.Name() == fieldName {
			return f.Value(), true
		}
	}

	return nil, false
}

// IsFieldPresent checks if a field was present in the input data
func (s *Schema) IsFieldPresent(fieldName string) bool {
	_, exists := s.presentFields[fieldName]
	return exists
}

// SetFieldPresent marks a field as present in the input data
func (s *Schema) SetFieldPresent(fieldName string) {
	s.presentFields[fieldName] = true
}

// WithSchema adds a field to a schema
func WithSchema(schema *Schema, field Field) {
	schema.fields = append(schema.fields, field)
}

// SubSchemaOption holds a callback for configuring sub-schemas
type SubSchemaOption[T any] struct {
	callback func(*Schema, *T)
}

// SubSchemaInterface is an interface for fields that can set sub-schema callbacks
type SubSchemaInterface[T any] interface {
	SetCallback(func(*Schema, *T))
}

// SubSchemaMapInterface is an interface for map fields that can set sub-schema callbacks
type SubSchemaMapInterface[K comparable, V any] interface {
	SetCallback(func(*Schema, K, V))
}

// Apply applies the sub-schema callback to the field
func (o SubSchemaOption[T]) Apply(field interface{}) {
	if f, ok := field.(SubSchemaInterface[T]); ok {
		f.SetCallback(o.callback)
	} else {
		panic(fmt.Sprintf("WithSubSchema doesn't support %T", field))
	}
}

// WithSubSchema creates a sub-schema option
func WithSubSchema[T any](callback func(*Schema, *T)) Option {
	return SubSchemaOption[T]{callback: callback}
}

// SubSchemaMapOption holds a callback for configuring map sub-schemas
type SubSchemaMapOption[K comparable, V any] struct {
	callback func(*Schema, K, V)
}

// Apply applies the map sub-schema callback to the field
func (o SubSchemaMapOption[K, V]) Apply(field interface{}) {
	if f, ok := field.(SubSchemaMapInterface[K, V]); ok {
		f.SetCallback(o.callback)
	} else {
		panic(fmt.Sprintf("WithSubSchemaMap doesn't support %T", field))
	}
}

// WithSubSchemaMap creates a map sub-schema option
func WithSubSchemaMap[K comparable, V any](callback func(*Schema, K, V)) Option {
	return SubSchemaMapOption[K, V]{callback: callback}
}
