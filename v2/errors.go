package poxxy

import (
	"fmt"
	"strings"
)

// ValidationError represents a single validation error.
type ValidationError struct {
	Field   string // Flat field name: "users[2].email"
	Value   any    // The invalid value
	Rule    string // "required", "min", "email", etc.
	Message string // Human-readable message
	Cause   error  // Underlying error
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

// newValidationError creates a ValidationError with template message interpolation.
func newValidationError(field string, value any, rule string, msg string, params map[string]any) *ValidationError {
	message := interpolateMessage(msg, field, value, params)
	return &ValidationError{
		Field:   field,
		Value:   value,
		Rule:    rule,
		Message: message,
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
