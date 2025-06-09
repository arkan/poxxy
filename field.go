package poxxy

import "fmt"

// FieldError represents a validation error for a specific field
type FieldError struct {
	Field string
	Error error
}

// Errors represents multiple validation errors
type Errors []FieldError

func (e Errors) Error() string {
	var msgs []string
	for _, err := range e {
		msgs = append(msgs, fmt.Sprintf("%s: %v", err.Field, err.Error))
	}
	// Manual join instead of using strings.Join
	if len(msgs) == 0 {
		return ""
	}
	result := msgs[0]
	for i := 1; i < len(msgs); i++ {
		result += "; " + msgs[i]
	}
	return result
}

// Field represents a field definition in a schema
type Field interface {
	Name() string
	Assign(data map[string]interface{}, schema *Schema) error
	Validate(schema *Schema) error
}
