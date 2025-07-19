package poxxy

import "fmt"

// FieldError represents a validation error for a specific field
type FieldError struct {
	Field       string
	Description string
	Error       error
}

// Errors represents multiple validation errors
type Errors []FieldError

// Error returns a string representation of all validation errors
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

// DescriptionOption holds a description
type DescriptionOption struct {
	description string
}

// Apply applies the description to the field
func (o DescriptionOption) Apply(field interface{}) {
	field.(Field).SetDescription(o.description)
}

// WithDescription creates a description option
func WithDescription(description string) Option {
	return DescriptionOption{description: description}
}

// Field represents a field definition in a schema
type Field interface {
	// Name returns the name of the field
	Name() string
	// Value returns the current value of the field
	Value() interface{}
	// Description returns the description of the field
	Description() string
	// SetDescription sets the description of the field
	SetDescription(description string)
	// Assign assigns a value to the field from the input data
	Assign(data map[string]interface{}, schema *Schema) error
	// Validate validates the field value using all registered validators
	Validate(schema *Schema) error
}
