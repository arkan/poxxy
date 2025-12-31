package poxxy

import (
	"fmt"
	"strings"
)

// ValidationError represents a single validation error.
type ValidationError struct {
	Field   string         // Flat field name: "users[2].email"
	Value   any            // The invalid value
	Rule    string         // "required", "min", "email", etc.
	Message string         // Human-readable message (already interpolated)
	Params  map[string]any // Parameters for message interpolation (for i18n)
	Cause   error          // Underlying error
}

func (e *ValidationError) Error() string {
	if e.Message != "" {
		return e.Message
	}
	if e.Cause != nil {
		return fmt.Sprintf("%s: %s", e.Field, e.Cause.Error())
	}
	return fmt.Sprintf("%s: validation failed (%s)", e.Field, e.Rule)
}

func (e *ValidationError) Unwrap() error {
	return e.Cause
}

// Translate translates the error message using the given catalog.
// Returns a new ValidationError with the translated message.
func (e *ValidationError) Translate(catalog *MessageCatalog) *ValidationError {
	if catalog == nil {
		return e
	}

	// Get translated message template from catalog
	msg := catalog.Get(MessageKey(e.Rule))

	// Interpolate with params
	translatedMsg := interpolateTemplate(msg, e.Params)

	return &ValidationError{
		Field:   e.Field,
		Value:   e.Value,
		Rule:    e.Rule,
		Message: translatedMsg,
		Params:  e.Params,
		Cause:   e.Cause,
	}
}

// ValidationErrors represents a collection of validation errors.
type ValidationErrors struct {
	Errors []*ValidationError
}

func (e *ValidationErrors) Error() string {
	if len(e.Errors) == 0 {
		return "validation failed"
	}
	if len(e.Errors) == 1 {
		return e.Errors[0].Error()
	}

	var sb strings.Builder
	sb.WriteString(fmt.Sprintf("%d validation errors:\n", len(e.Errors)))
	for _, err := range e.Errors {
		sb.WriteString("  - ")
		sb.WriteString(err.Error())
		sb.WriteString("\n")
	}
	return sb.String()
}

func (e *ValidationErrors) Add(err *ValidationError) {
	e.Errors = append(e.Errors, err)
}

func (e *ValidationErrors) HasErrors() bool {
	return len(e.Errors) > 0
}

// Merge adds all errors from another ValidationErrors.
func (e *ValidationErrors) Merge(other *ValidationErrors) {
	if other == nil {
		return
	}
	e.Errors = append(e.Errors, other.Errors...)
}

// Translate translates all error messages using the given catalog.
// Returns a new ValidationErrors with translated messages.
func (e *ValidationErrors) Translate(catalog *MessageCatalog) *ValidationErrors {
	if catalog == nil || len(e.Errors) == 0 {
		return e
	}

	translated := &ValidationErrors{
		Errors: make([]*ValidationError, len(e.Errors)),
	}
	for i, err := range e.Errors {
		translated.Errors[i] = err.Translate(catalog)
	}
	return translated
}

// newValidationError creates a ValidationError with template message interpolation.
func newValidationError(field string, value any, rule string, msg string, params map[string]any) *ValidationError {
	// Merge field and value into params for later translation
	allParams := make(map[string]any)
	for k, v := range params {
		allParams[k] = v
	}
	allParams["field"] = field
	allParams["value"] = value

	message := interpolateMessage(msg, field, value, params)
	return &ValidationError{
		Field:   field,
		Value:   value,
		Rule:    rule,
		Message: message,
		Params:  allParams,
	}
}

// interpolateMessage replaces placeholders in message templates.
// Supported: {field}, {value}, {min}, {max}, {expected}, {error}
func interpolateMessage(msg string, field string, value any, params map[string]any) string {
	if msg == "" {
		return ""
	}

	result := msg
	result = strings.ReplaceAll(result, "{field}", field)
	result = strings.ReplaceAll(result, "{value}", fmt.Sprintf("%v", value))

	for key, val := range params {
		placeholder := "{" + key + "}"
		result = strings.ReplaceAll(result, placeholder, fmt.Sprintf("%v", val))
	}

	return result
}
