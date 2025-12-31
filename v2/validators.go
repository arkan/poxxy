package poxxy

import (
	"fmt"
	"reflect"
	"regexp"
	"strings"

	"golang.org/x/text/cases"
	"golang.org/x/text/language"
)

// baseValidator provides common functionality for all validators.
type baseValidator[T any] struct {
	rule    string
	msg     string
	params  map[string]any
	checkFn func(value T, field string) error
}

func (v *baseValidator[T]) Validate(value T, field string) error {
	err := v.checkFn(value, field)
	if err != nil && v.msg != "" {
		return newValidationError(field, value, v.rule, v.msg, v.params)
	}
	return err
}

func (v *baseValidator[T]) Rule() string {
	return v.rule
}

func (v *baseValidator[T]) WithMessage(msg string) Validator[T] {
	return &baseValidator[T]{
		rule:    v.rule,
		msg:     msg,
		params:  v.params,
		checkFn: v.checkFn,
	}
}

// =============================================================================
// Required Validators
// =============================================================================

// Required validates that a value is not its zero value.
func Required[T any]() Validator[T] {
	return &baseValidator[T]{
		rule:   "required",
		params: nil,
		checkFn: func(value T, field string) error {
			if isZeroValue(value) {
				return newValidationError(field, value, "required", "{field} is required", nil)
			}
			return nil
		},
	}
}

// NotEmpty validates that a value is not empty (for strings, slices, maps).
func NotEmpty[T any]() Validator[T] {
	return &baseValidator[T]{
		rule:   "not_empty",
		params: nil,
		checkFn: func(value T, field string) error {
			if isEmpty(value) {
				return newValidationError(field, value, "not_empty", "{field} must not be empty", nil)
			}
			return nil
		},
	}
}

// =============================================================================
// String Validators
// =============================================================================

var emailRegex = regexp.MustCompile(`^[a-zA-Z0-9._%+\-]+@[a-zA-Z0-9.\-]+\.[a-zA-Z]{2,}$`)

// Email validates that a string is a valid email address.
func Email() Validator[string] {
	return &baseValidator[string]{
		rule:   "email",
		params: nil,
		checkFn: func(value string, field string) error {
			if value == "" {
				return nil // Use Required() for non-empty check
			}
			if !emailRegex.MatchString(value) {
				return newValidationError(field, value, "email", "{field} must be a valid email address", nil)
			}
			return nil
		},
	}
}

var urlRegex = regexp.MustCompile(`^https?://[^\s/$.?#].[^\s]*$`)

// URL validates that a string is a valid URL.
func URL() Validator[string] {
	return &baseValidator[string]{
		rule:   "url",
		params: nil,
		checkFn: func(value string, field string) error {
			if value == "" {
				return nil
			}
			if !urlRegex.MatchString(value) {
				return newValidationError(field, value, "url", "{field} must be a valid URL", nil)
			}
			return nil
		},
	}
}

var uuidRegex = regexp.MustCompile(`^[0-9a-fA-F]{8}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{4}-[0-9a-fA-F]{12}$`)

// UUID validates that a string is a valid UUID (v1-v5).
func UUID() Validator[string] {
	return &baseValidator[string]{
		rule:   "uuid",
		params: nil,
		checkFn: func(value string, field string) error {
			if value == "" {
				return nil
			}
			if !uuidRegex.MatchString(value) {
				return newValidationError(field, value, "uuid", "{field} must be a valid UUID", nil)
			}
			return nil
		},
	}
}

// Pattern validates that a string matches a regex pattern.
func Pattern(pattern string) Validator[string] {
	re := regexp.MustCompile(pattern)
	return &baseValidator[string]{
		rule:   "pattern",
		params: map[string]any{"pattern": pattern},
		checkFn: func(value string, field string) error {
			if value == "" {
				return nil
			}
			if !re.MatchString(value) {
				return newValidationError(field, value, "pattern", "{field} must match pattern {pattern}", map[string]any{"pattern": pattern})
			}
			return nil
		},
	}
}

// =============================================================================
// Numeric Validators
// =============================================================================

// Ordered is a constraint for types that support < > operators.
type Ordered interface {
	~int | ~int8 | ~int16 | ~int32 | ~int64 |
		~uint | ~uint8 | ~uint16 | ~uint32 | ~uint64 |
		~float32 | ~float64 | ~string
}

// Min validates that a value is >= the minimum.
func Min[T Ordered](min T) Validator[T] {
	return &baseValidator[T]{
		rule:   "min",
		params: map[string]any{"min": min},
		checkFn: func(value T, field string) error {
			if value < min {
				return newValidationError(field, value, "min", "{field} must be at least {min}", map[string]any{"min": min})
			}
			return nil
		},
	}
}

// Max validates that a value is <= the maximum.
func Max[T Ordered](max T) Validator[T] {
	return &baseValidator[T]{
		rule:   "max",
		params: map[string]any{"max": max},
		checkFn: func(value T, field string) error {
			if value > max {
				return newValidationError(field, value, "max", "{field} must be at most {max}", map[string]any{"max": max})
			}
			return nil
		},
	}
}

// Between validates that a value is within a range (inclusive).
func Between[T Ordered](min, max T) Validator[T] {
	return &baseValidator[T]{
		rule:   "between",
		params: map[string]any{"min": min, "max": max},
		checkFn: func(value T, field string) error {
			if value < min || value > max {
				return newValidationError(field, value, "between", "{field} must be between {min} and {max}", map[string]any{"min": min, "max": max})
			}
			return nil
		},
	}
}

// =============================================================================
// Length Validators
// =============================================================================

// Lengthable is a constraint for types that have a length.
type Lengthable interface {
	~string | ~[]byte
}

// MinLength validates that a string or []byte has at least n characters/bytes.
func MinLength[T Lengthable](n int) Validator[T] {
	return &baseValidator[T]{
		rule:   "min_length",
		params: map[string]any{"min": n},
		checkFn: func(value T, field string) error {
			if len(value) < n {
				return newValidationError(field, value, "min_length", "{field} must be at least {min} characters", map[string]any{"min": n})
			}
			return nil
		},
	}
}

// MaxLength validates that a string or []byte has at most n characters/bytes.
func MaxLength[T Lengthable](n int) Validator[T] {
	return &baseValidator[T]{
		rule:   "max_length",
		params: map[string]any{"max": n},
		checkFn: func(value T, field string) error {
			if len(value) > n {
				return newValidationError(field, value, "max_length", "{field} must be at most {max} characters", map[string]any{"max": n})
			}
			return nil
		},
	}
}

// Length validates that a string or []byte has exactly n characters/bytes.
func Length[T Lengthable](n int) Validator[T] {
	return &baseValidator[T]{
		rule:   "length",
		params: map[string]any{"length": n},
		checkFn: func(value T, field string) error {
			if len(value) != n {
				return newValidationError(field, value, "length", "{field} must be exactly {length} characters", map[string]any{"length": n})
			}
			return nil
		},
	}
}

// =============================================================================
// Slice Validators
// =============================================================================

// MinItems validates that a slice has at least n items.
func MinItems[T any](n int) Validator[[]T] {
	return &baseValidator[[]T]{
		rule:   "min_items",
		params: map[string]any{"min": n},
		checkFn: func(value []T, field string) error {
			if len(value) < n {
				return newValidationError(field, value, "min_items", "{field} must have at least {min} items", map[string]any{"min": n})
			}
			return nil
		},
	}
}

// MaxItems validates that a slice has at most n items.
func MaxItems[T any](n int) Validator[[]T] {
	return &baseValidator[[]T]{
		rule:   "max_items",
		params: map[string]any{"max": n},
		checkFn: func(value []T, field string) error {
			if len(value) > n {
				return newValidationError(field, value, "max_items", "{field} must have at most {max} items", map[string]any{"max": n})
			}
			return nil
		},
	}
}

// Each validates each item in a slice with the given validators.
func Each[T any](validators ...Validator[T]) Validator[[]T] {
	return &eachValidator[T]{
		validators: validators,
	}
}

type eachValidator[T any] struct {
	validators []Validator[T]
	msg        string
}

func (v *eachValidator[T]) Validate(value []T, field string) error {
	errs := &ValidationErrors{}
	for i, item := range value {
		itemField := fmt.Sprintf("%s[%d]", field, i)
		for _, validator := range v.validators {
			if err := validator.Validate(item, itemField); err != nil {
				if ve, ok := err.(*ValidationError); ok {
					errs.Add(ve)
				} else {
					errs.Add(&ValidationError{
						Field:   itemField,
						Value:   item,
						Rule:    validator.Rule(),
						Message: err.Error(),
						Cause:   err,
					})
				}
			}
		}
	}
	if errs.HasErrors() {
		return errs.Errors[0] // Return first error
	}
	return nil
}

func (v *eachValidator[T]) Rule() string {
	return "each"
}

func (v *eachValidator[T]) WithMessage(msg string) Validator[[]T] {
	return &eachValidator[T]{
		validators: v.validators,
		msg:        msg,
	}
}

// Unique validates that all items in a slice are unique.
func Unique[T comparable]() Validator[[]T] {
	return &baseValidator[[]T]{
		rule:   "unique",
		params: nil,
		checkFn: func(value []T, field string) error {
			seen := make(map[T]int)
			for i, item := range value {
				if firstIdx, exists := seen[item]; exists {
					return newValidationError(field, value, "unique",
						fmt.Sprintf("{field} contains duplicate value at index %d and %d", firstIdx, i), nil)
				}
				seen[item] = i
			}
			return nil
		},
	}
}

// UniqueBy validates that all items in a slice are unique by a key function.
func UniqueBy[T any, K comparable](keyFn func(T) K) Validator[[]T] {
	return &baseValidator[[]T]{
		rule:   "unique_by",
		params: nil,
		checkFn: func(value []T, field string) error {
			seen := make(map[K]int)
			for i, item := range value {
				key := keyFn(item)
				if firstIdx, exists := seen[key]; exists {
					return newValidationError(field, value, "unique_by",
						fmt.Sprintf("{field} contains duplicate key at index %d and %d", firstIdx, i), nil)
				}
				seen[key] = i
			}
			return nil
		},
	}
}

// =============================================================================
// Enum Validators
// =============================================================================

// In validates that a value is one of the allowed values.
func In[T comparable](allowed ...T) Validator[T] {
	allowedSet := make(map[T]struct{}, len(allowed))
	for _, v := range allowed {
		allowedSet[v] = struct{}{}
	}

	return &baseValidator[T]{
		rule:   "in",
		params: map[string]any{"allowed": allowed},
		checkFn: func(value T, field string) error {
			if _, ok := allowedSet[value]; !ok {
				return newValidationError(field, value, "in", "{field} must be one of the allowed values", map[string]any{"allowed": allowed})
			}
			return nil
		},
	}
}

// NotIn validates that a value is not one of the disallowed values.
func NotIn[T comparable](disallowed ...T) Validator[T] {
	disallowedSet := make(map[T]struct{}, len(disallowed))
	for _, v := range disallowed {
		disallowedSet[v] = struct{}{}
	}

	return &baseValidator[T]{
		rule:   "not_in",
		params: map[string]any{"disallowed": disallowed},
		checkFn: func(value T, field string) error {
			if _, ok := disallowedSet[value]; ok {
				return newValidationError(field, value, "not_in", "{field} must not be one of the disallowed values", map[string]any{"disallowed": disallowed})
			}
			return nil
		},
	}
}

// =============================================================================
// Cross-Field Validators
// =============================================================================

// RequiredIf validates that a field is required if a condition is true.
// The condition function receives the schema to access other field values.
func RequiredIf[T any](condition func(*Schema) bool) Validator[T] {
	return &requiredIfValidator[T]{
		condition: condition,
	}
}

type requiredIfValidator[T any] struct {
	condition func(*Schema) bool
	schema    *Schema
	msg       string
}

func (v *requiredIfValidator[T]) Validate(value T, field string) error {
	// Note: The schema context is set during validation phase
	// For now, we check if the value is zero
	if v.schema != nil && v.condition(v.schema) {
		if isZeroValue(value) {
			msg := v.msg
			if msg == "" {
				msg = "{field} is required"
			}
			return newValidationError(field, value, "required_if", msg, nil)
		}
	}
	return nil
}

func (v *requiredIfValidator[T]) Rule() string {
	return "required_if"
}

func (v *requiredIfValidator[T]) WithMessage(msg string) Validator[T] {
	return &requiredIfValidator[T]{
		condition: v.condition,
		schema:    v.schema,
		msg:       msg,
	}
}

// SetSchema sets the schema context for cross-field validation.
func (v *requiredIfValidator[T]) SetSchema(s *Schema) {
	v.schema = s
}

// =============================================================================
// Custom Validator
// =============================================================================

// ValidatorFunc creates a validator from a function.
func ValidatorFunc[T any](fn func(T) error) Validator[T] {
	return &funcValidator[T]{
		fn: fn,
	}
}

type funcValidator[T any] struct {
	fn  func(T) error
	msg string
}

func (v *funcValidator[T]) Validate(value T, field string) error {
	err := v.fn(value)
	if err != nil {
		if v.msg != "" {
			return newValidationError(field, value, "custom", v.msg, nil)
		}
		return newValidationError(field, value, "custom", err.Error(), nil)
	}
	return nil
}

func (v *funcValidator[T]) Rule() string {
	return "custom"
}

func (v *funcValidator[T]) WithMessage(msg string) Validator[T] {
	return &funcValidator[T]{
		fn:  v.fn,
		msg: msg,
	}
}

// =============================================================================
// Deferred Validators (Async Validation)
// =============================================================================

// DeferredCheck is a function that performs deferred validation (e.g., DB lookups).
type DeferredCheck func() error

// DeferredValidator is a validator that returns a deferred check function.
// Use this for validations requiring external calls (DB, API, etc.).
type DeferredValidator[T any] interface {
	Validate(value T, field string) (DeferredCheck, error)
	Rule() string
	WithMessage(msg string) DeferredValidator[T]
}

// deferredValidator wraps a function as a deferred validator.
type deferredValidator[T any] struct {
	rule string
	msg  string
	fn   func(value T, field string) DeferredCheck
}

// Deferred creates a deferred validator from a function.
// The function returns a DeferredCheck that will be executed after synchronous validation.
func Deferred[T any](fn func(value T, field string) DeferredCheck) DeferredValidator[T] {
	return &deferredValidator[T]{
		rule: "deferred",
		fn:   fn,
	}
}

// DeferredNamed creates a named deferred validator from a function.
func DeferredNamed[T any](rule string, fn func(value T, field string) DeferredCheck) DeferredValidator[T] {
	return &deferredValidator[T]{
		rule: rule,
		fn:   fn,
	}
}

func (v *deferredValidator[T]) Validate(value T, field string) (DeferredCheck, error) {
	check := v.fn(value, field)
	return check, nil
}

func (v *deferredValidator[T]) Rule() string {
	return v.rule
}

func (v *deferredValidator[T]) WithMessage(msg string) DeferredValidator[T] {
	return &deferredValidator[T]{
		rule: v.rule,
		msg:  msg,
		fn:   v.fn,
	}
}

// DeferredChecks collects deferred checks for batch execution.
type DeferredChecks struct {
	checks []struct {
		field string
		check DeferredCheck
	}
}

// Add adds a deferred check to the collection.
func (d *DeferredChecks) Add(field string, check DeferredCheck) {
	if check != nil {
		d.checks = append(d.checks, struct {
			field string
			check DeferredCheck
		}{field: field, check: check})
	}
}

// Run executes all deferred checks sequentially and returns any errors.
func (d *DeferredChecks) Run() *ValidationErrors {
	errs := &ValidationErrors{}
	for _, c := range d.checks {
		if err := c.check(); err != nil {
			errs.Add(&ValidationError{
				Field:   c.field,
				Rule:    "deferred",
				Message: err.Error(),
				Cause:   err,
			})
		}
	}
	if errs.HasErrors() {
		return errs
	}
	return nil
}

// RunParallel executes all deferred checks in parallel and returns any errors.
func (d *DeferredChecks) RunParallel() *ValidationErrors {
	if len(d.checks) == 0 {
		return nil
	}

	type result struct {
		field string
		err   error
	}

	results := make(chan result, len(d.checks))

	for _, c := range d.checks {
		go func(field string, check DeferredCheck) {
			results <- result{field: field, err: check()}
		}(c.field, c.check)
	}

	errs := &ValidationErrors{}
	for range d.checks {
		r := <-results
		if r.err != nil {
			errs.Add(&ValidationError{
				Field:   r.field,
				Rule:    "deferred",
				Message: r.err.Error(),
				Cause:   r.err,
			})
		}
	}

	if errs.HasErrors() {
		return errs
	}
	return nil
}

// HasChecks returns true if there are deferred checks to run.
func (d *DeferredChecks) HasChecks() bool {
	return len(d.checks) > 0
}

// =============================================================================
// Helper Functions
// =============================================================================

// isZeroValue checks if a value is the zero value for its type.
func isZeroValue[T any](v T) bool {
	return reflect.ValueOf(&v).Elem().IsZero()
}

// isEmpty checks if a value is empty (zero length for strings, slices, maps).
func isEmpty[T any](v T) bool {
	rv := reflect.ValueOf(&v).Elem()
	switch rv.Kind() {
	case reflect.String, reflect.Slice, reflect.Map, reflect.Array:
		return rv.Len() == 0
	default:
		return rv.IsZero()
	}
}

// Unused but kept for potential i18n support
var _ = cases.Title(language.English)
var _ = strings.TrimSpace
