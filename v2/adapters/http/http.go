// Package http provides HTTP request adapters for poxxy schemas.
// These adapters convert HTTP requests into the format expected by poxxy.Schema.ParseDecoder().
package http

import (
	"encoding/json"
	"fmt"
	"io"
	"mime"
	"net/http"
	"net/url"
	"strconv"
	"strings"
)

// MaxBodySize is the default maximum request body size (5MB).
var MaxBodySize int64 = 5 << 20

// ContentType represents supported content types.
type ContentType int

const (
	ContentTypeAuto ContentType = iota
	ContentTypeJSON
	ContentTypeForm
	ContentTypeQuery
)

// Options configures HTTP request parsing.
type Options struct {
	MaxBodySize int64
	ContentType ContentType
}

// DefaultOptions returns the default options.
func DefaultOptions() *Options {
	return &Options{
		MaxBodySize: MaxBodySize,
		ContentType: ContentTypeAuto,
	}
}

// =============================================================================
// JSON Decoder
// =============================================================================

// jsonDecoder wraps an io.Reader for JSON decoding.
type jsonDecoder struct {
	reader io.Reader
}

// JSONDecoder creates a decoder for JSON content.
func JSONDecoder(r io.Reader) *jsonDecoder {
	return &jsonDecoder{reader: r}
}

// JSONDecoderFromRequest creates a JSON decoder from an HTTP request.
func JSONDecoderFromRequest(r *http.Request, opts *Options) *jsonDecoder {
	if opts == nil {
		opts = DefaultOptions()
	}
	return &jsonDecoder{
		reader: io.LimitReader(r.Body, opts.MaxBodySize),
	}
}

func (d *jsonDecoder) Decode(v any) error {
	return json.NewDecoder(d.reader).Decode(v)
}

// =============================================================================
// Form Decoder (application/x-www-form-urlencoded)
// =============================================================================

// formDecoder decodes form data.
type formDecoder struct {
	values url.Values
}

// FormDecoder creates a decoder from url.Values.
func FormDecoder(values url.Values) *formDecoder {
	return &formDecoder{values: values}
}

// FormDecoderFromRequest creates a form decoder from an HTTP request.
// It parses the request body as application/x-www-form-urlencoded.
func FormDecoderFromRequest(r *http.Request, opts *Options) (*formDecoder, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	// Limit body size
	r.Body = http.MaxBytesReader(nil, r.Body, opts.MaxBodySize)

	if err := r.ParseForm(); err != nil {
		return nil, fmt.Errorf("failed to parse form: %w", err)
	}

	return &formDecoder{values: r.PostForm}, nil
}

func (d *formDecoder) Decode(v any) error {
	target, ok := v.(*map[string]any)
	if !ok {
		return fmt.Errorf("form decoder expects *map[string]any, got %T", v)
	}

	result := make(map[string]any)
	for key, values := range d.values {
		if len(values) == 1 {
			result[key] = convertFormValue(values[0])
		} else {
			converted := make([]any, len(values))
			for i, val := range values {
				converted[i] = convertFormValue(val)
			}
			result[key] = converted
		}
	}

	*target = result
	return nil
}

// convertFormValue attempts to convert a form string value to a more specific type.
func convertFormValue(s string) any {
	// Empty string
	if s == "" {
		return s
	}

	// Try boolean
	if s == "true" {
		return true
	}
	if s == "false" {
		return false
	}

	// Try integer
	if i, err := strconv.ParseInt(s, 10, 64); err == nil {
		// Check if it fits in int
		if i >= -2147483648 && i <= 2147483647 {
			return int(i)
		}
		return i
	}

	// Try float
	if f, err := strconv.ParseFloat(s, 64); err == nil {
		return f
	}

	// Return as string
	return s
}

// =============================================================================
// Query Decoder (URL query parameters)
// =============================================================================

// queryDecoder decodes URL query parameters.
type queryDecoder struct {
	values url.Values
}

// QueryDecoder creates a decoder from url.Values.
func QueryDecoder(values url.Values) *queryDecoder {
	return &queryDecoder{values: values}
}

// QueryDecoderFromRequest creates a query decoder from an HTTP request.
func QueryDecoderFromRequest(r *http.Request) *queryDecoder {
	return &queryDecoder{values: r.URL.Query()}
}

func (d *queryDecoder) Decode(v any) error {
	target, ok := v.(*map[string]any)
	if !ok {
		return fmt.Errorf("query decoder expects *map[string]any, got %T", v)
	}

	result := make(map[string]any)
	for key, values := range d.values {
		if len(values) == 1 {
			result[key] = convertFormValue(values[0])
		} else {
			converted := make([]any, len(values))
			for i, val := range values {
				converted[i] = convertFormValue(val)
			}
			result[key] = converted
		}
	}

	*target = result
	return nil
}

// =============================================================================
// Multipart Form Decoder
// =============================================================================

// multipartDecoder decodes multipart form data.
type multipartDecoder struct {
	values url.Values
}

// MultipartDecoderFromRequest creates a multipart form decoder from an HTTP request.
func MultipartDecoderFromRequest(r *http.Request, opts *Options) (*multipartDecoder, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	if err := r.ParseMultipartForm(opts.MaxBodySize); err != nil {
		return nil, fmt.Errorf("failed to parse multipart form: %w", err)
	}

	return &multipartDecoder{values: r.MultipartForm.Value}, nil
}

func (d *multipartDecoder) Decode(v any) error {
	target, ok := v.(*map[string]any)
	if !ok {
		return fmt.Errorf("multipart decoder expects *map[string]any, got %T", v)
	}

	result := make(map[string]any)
	for key, values := range d.values {
		if len(values) == 1 {
			result[key] = convertFormValue(values[0])
		} else {
			converted := make([]any, len(values))
			for i, val := range values {
				converted[i] = convertFormValue(val)
			}
			result[key] = converted
		}
	}

	*target = result
	return nil
}

// =============================================================================
// Auto-Detect Decoder
// =============================================================================

// Decoder is the interface that decoders must implement.
type Decoder interface {
	Decode(v any) error
}

// DecoderFromRequest creates a decoder based on the request's Content-Type header.
// It automatically detects JSON, form-urlencoded, and multipart form data.
// For GET requests without a body, it uses query parameters.
func DecoderFromRequest(r *http.Request, opts *Options) (Decoder, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	// Override content type if specified
	if opts.ContentType != ContentTypeAuto {
		switch opts.ContentType {
		case ContentTypeJSON:
			return JSONDecoderFromRequest(r, opts), nil
		case ContentTypeForm:
			return FormDecoderFromRequest(r, opts)
		case ContentTypeQuery:
			return QueryDecoderFromRequest(r), nil
		}
	}

	// For GET/HEAD requests without body, use query parameters
	if r.Method == http.MethodGet || r.Method == http.MethodHead {
		return QueryDecoderFromRequest(r), nil
	}

	// Parse Content-Type header
	contentType := r.Header.Get("Content-Type")
	if contentType == "" {
		// Default to JSON if no content type specified
		return JSONDecoderFromRequest(r, opts), nil
	}

	mediaType, _, err := mime.ParseMediaType(contentType)
	if err != nil {
		return nil, fmt.Errorf("invalid Content-Type header: %w", err)
	}

	switch mediaType {
	case "application/json":
		return JSONDecoderFromRequest(r, opts), nil
	case "application/x-www-form-urlencoded":
		return FormDecoderFromRequest(r, opts)
	case "multipart/form-data":
		return MultipartDecoderFromRequest(r, opts)
	default:
		return nil, fmt.Errorf("unsupported Content-Type: %s", mediaType)
	}
}

// =============================================================================
// Nested Form Parsing (field[key][subfield])
// =============================================================================

// nestedFormDecoder handles nested form data like field[key][subfield]=value.
type nestedFormDecoder struct {
	values url.Values
}

// NestedFormDecoder creates a decoder that handles nested form syntax.
func NestedFormDecoder(values url.Values) *nestedFormDecoder {
	return &nestedFormDecoder{values: values}
}

// NestedFormDecoderFromRequest creates a nested form decoder from an HTTP request.
func NestedFormDecoderFromRequest(r *http.Request, opts *Options) (*nestedFormDecoder, error) {
	if opts == nil {
		opts = DefaultOptions()
	}

	r.Body = http.MaxBytesReader(nil, r.Body, opts.MaxBodySize)

	if err := r.ParseForm(); err != nil {
		return nil, fmt.Errorf("failed to parse form: %w", err)
	}

	return &nestedFormDecoder{values: r.PostForm}, nil
}

func (d *nestedFormDecoder) Decode(v any) error {
	target, ok := v.(*map[string]any)
	if !ok {
		return fmt.Errorf("nested form decoder expects *map[string]any, got %T", v)
	}

	result := make(map[string]any)

	for key, values := range d.values {
		value := values[0]
		if len(values) > 1 {
			// Handle multiple values as array
			arr := make([]any, len(values))
			for i, v := range values {
				arr[i] = convertFormValue(v)
			}
			setNestedValue(result, key, arr)
		} else {
			setNestedValue(result, key, convertFormValue(value))
		}
	}

	*target = result
	return nil
}

// setNestedValue sets a value in a nested map structure.
// Supports syntax like: field[key][subfield] or field[0][name]
func setNestedValue(m map[string]any, key string, value any) {
	parts := parseNestedKey(key)
	if len(parts) == 0 {
		return
	}

	current := m
	for i := 0; i < len(parts)-1; i++ {
		part := parts[i]

		if existing, ok := current[part]; ok {
			if nested, ok := existing.(map[string]any); ok {
				current = nested
			} else {
				// Overwrite non-map value
				nested := make(map[string]any)
				current[part] = nested
				current = nested
			}
		} else {
			nested := make(map[string]any)
			current[part] = nested
			current = nested
		}
	}

	current[parts[len(parts)-1]] = value
}

// parseNestedKey parses a key like "field[key][subfield]" into ["field", "key", "subfield"].
func parseNestedKey(key string) []string {
	if !strings.Contains(key, "[") {
		return []string{key}
	}

	var parts []string
	var current strings.Builder

	inBracket := false
	for _, r := range key {
		switch r {
		case '[':
			if current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
			inBracket = true
		case ']':
			if inBracket && current.Len() > 0 {
				parts = append(parts, current.String())
				current.Reset()
			}
			inBracket = false
		default:
			current.WriteRune(r)
		}
	}

	if current.Len() > 0 {
		parts = append(parts, current.String())
	}

	return parts
}
