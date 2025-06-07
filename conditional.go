package poxxy

// ConditionalValidator represents a conditional validation rule
type ConditionalValidator struct {
	condition Condition
	option    Option
}

// Condition represents a validation condition
type Condition interface {
	Check(schema *Schema) bool
}

// FieldEqualsCondition checks if a field equals a specific value
type FieldEqualsCondition struct {
	fieldName string
	value     interface{}
}

func (c *FieldEqualsCondition) Check(schema *Schema) bool {
	// For now, this is a placeholder - full implementation would require
	// access to the current field values during validation
	// This is a simplified version for the API demonstration
	return true
}

// FieldEquals creates a condition that checks if a field equals a value
func FieldEquals(fieldName string, value interface{}) Condition {
	return &FieldEqualsCondition{
		fieldName: fieldName,
		value:     value,
	}
}

// ConditionalValidators creates conditional validation options
func ConditionalValidators(condition Condition, option Option) Option {
	return &ConditionalValidator{
		condition: condition,
		option:    option,
	}
}

func (cv *ConditionalValidator) Apply(field interface{}) {
	// For demonstration purposes, we'll apply the validators unconditionally
	// A full implementation would need to track field values and apply conditionally
	cv.option.Apply(field)
}
